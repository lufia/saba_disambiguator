package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	sabadisambiguator "saba_disambiguator/lib"

	"cloud.google.com/go/bigquery"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dghubble/go-twitter/twitter"
	"google.golang.org/api/option"
)

type ItemForBigQuery struct {
	CreatedAt  time.Time `bigquery:"created_at"`
	IdStr      string    `bigquery:"id_str"`
	ScreenName string    `bigquery:"screen_name"`
	Text       string    `bigquery:"text"`
	RawJson    string    `bigquery:"raw_json"`
	Score      float64   `bigquery:"score"`
	IsPositive bool      `bigquery:"is_positive"`
}

type SlackConfig struct {
	WebhookUrlPositive string
	WebhookUrlNegative string
}

func getSlackConfig(svc *ssm.SSM, config sabadisambiguator.Config) (SlackConfig, error) {
	slackConfig := SlackConfig{}

	webhookUrlPositive, err := sabadisambiguator.GetValueFromParameterStore(svc, config.SlackConfig.ParameterStoreNameWebhookUrlPositive)
	if err != nil {
		return slackConfig, err
	}
	slackConfig.WebhookUrlPositive = webhookUrlPositive

	webhookUrlNegative, err := sabadisambiguator.GetValueFromParameterStore(svc, config.SlackConfig.ParameterStoreNameWebhookUrlNegative)
	if err != nil {
		return slackConfig, err
	}
	slackConfig.WebhookUrlNegative = webhookUrlNegative

	return slackConfig, nil
}

func DoDisambiguate() error {
	config, err := sabadisambiguator.GetConfigFromFile("config.yml")
	if err != nil {
		return err
	}
	svc := ssm.New(session.New(), &aws.Config{
		Region: aws.String(config.Region),
	})

	client, err := sabadisambiguator.GetTwitterClient(svc, *config)
	if err != nil {
		return err
	}

	slackConfig, err := getSlackConfig(svc, *config)
	if err != nil {
		return err
	}

	model, err := sabadisambiguator.LoadPerceptron("model.bin")
	if err != nil {
		return err
	}
	query := "mackerel lang:ja exclude:retweets"
	if config.Query != "" {
		query = config.Query
	}

	search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query:      query,
		Count:      100,
		ResultType: "recent",
	})

	if err != nil {
		return err
	}

	now := time.Now()
	itemsForBq := make([]*ItemForBigQuery, 0)

	for _, t := range search.Statuses {
		createdAt, err := t.CreatedAtTime()
		if err != nil {
			return err
		}
		if now.After(createdAt.Add(5 * time.Minute)) {
			continue
		}

		tweetJson, err := json.Marshal(t)
		if err != nil {
			return err
		}
		fv := sabadisambiguator.ExtractFeaturesWithOptions(t, sabadisambiguator.ExtractOptions{
			ScreenNames: config.ScreenNames,
		})
		score := model.PredictScore(fv)
		predLabel := model.Predict(fv)
		item := ItemForBigQuery{
			CreatedAt:  createdAt,
			IdStr:      t.IDStr,
			ScreenName: t.User.ScreenName,
			Text:       t.Text,
			RawJson:    string(tweetJson),
			Score:      score,
			IsPositive: predLabel == sabadisambiguator.POSITIVE,
		}
		itemsForBq = append(itemsForBq, &item)

		tweetPermalink := fmt.Sprintf("https://twitter.com/%s/status/%s", t.User.ScreenName, t.IDStr)
		payload := slack.Payload{
			Text: tweetPermalink,
		}
		fmt.Println(tweetPermalink)

		if predLabel == sabadisambiguator.POSITIVE {
			fmt.Fprintf(os.Stderr, "%s\n", tweetPermalink)
			err := slack.Send(slackConfig.WebhookUrlPositive, "", payload)
			if err != nil {
				return err[0]
			}
		} else if (predLabel == sabadisambiguator.NEGATIVE) && (slackConfig.WebhookUrlNegative != "") {
			err := slack.Send(slackConfig.WebhookUrlNegative, "", payload)
			if err != nil {
				return err[0]
			}
		}
	}

	if config.BigQueryConfig.ProjectId != "" && len(itemsForBq) > 0 {
		serviceAccountCredential, err := sabadisambiguator.GetValueFromParameterStore(svc, config.BigQueryConfig.ParameterStoreNameServiceAccountCredential)
		if err != nil {
			return err
		}
		ctx := context.Background()
		bqClient, err := bigquery.NewClient(ctx, config.BigQueryConfig.ProjectId, option.WithCredentialsJSON([]byte(serviceAccountCredential)))
		if err != nil {
			return err
		}
		defer bqClient.Close()

		u := bqClient.Dataset(config.BigQueryConfig.Dataset).Table(config.BigQueryConfig.Table).Inserter()
		err = u.Put(ctx, itemsForBq)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(DoDisambiguate)
}
