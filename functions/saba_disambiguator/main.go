package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"

	twitter2 "github.com/syou6162/saba_disambiguator/twitter"

	"cloud.google.com/go/bigquery"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/syou6162/saba_disambiguator/slack"
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
	query := "mackerel lang:ja -is:retweet"
	if config.Query != "" {
		query = config.Query
	}

	resp, err := client.RecentSearch(query)

	if err != nil {
		return err
	}

	now := time.Now()
	itemsForBq := make([]*ItemForBigQuery, 0)

	for _, t := range resp {
		createdAt := t.CreatedAt
		if now.After(createdAt.Add(5 * time.Minute)) {
			continue
		}
		if isSpam(t, config.SpamTexts) {
			continue
		}

		fv := sabadisambiguator.ExtractFeaturesWithOptions(t, sabadisambiguator.ExtractOptions{
			ScreenNames: config.ScreenNames,
		})
		score := model.PredictScore(fv)
		predLabel := model.Predict(fv)
		tweetJson, err := json.Marshal(t)
		if err != nil {
			return err
		}
		item := ItemForBigQuery{
			CreatedAt:  createdAt,
			IdStr:      t.ID,
			ScreenName: t.User.UserName,
			Text:       t.Text,
			RawJson:    string(tweetJson),
			Score:      score,
			IsPositive: predLabel == sabadisambiguator.POSITIVE,
		}
		itemsForBq = append(itemsForBq, &item)

		payload := formatTweetIntoSlackPayload(t)

		err = nil
		if predLabel == sabadisambiguator.POSITIVE {
			err = postJSON(slackConfig.WebhookUrlPositive, payload)
		} else if (predLabel == sabadisambiguator.NEGATIVE) && (slackConfig.WebhookUrlNegative != "") {
			err = postJSON(slackConfig.WebhookUrlNegative, payload)
		}
		if err != nil {
			return err
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

func postJSON(url string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("postjson: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("postjson: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("postjson: failed with status: %s", resp.Status)
	}

	return nil
}

func isSpam(t *twitter2.Tweet, texts []string) bool {
	return slices.ContainsFunc(texts, func(s string) bool {
		return strings.Contains(t.Text, s)
	})
}

func formatTweetIntoSlackPayload(t *twitter2.Tweet) slack.Payload {
	permalink := fmt.Sprintf("https://twitter.com/%s/status/%s", t.User.UserName, t.ID)

	return slack.Payload{
		Text: permalink,
		Blocks: []any{
			slack.ContextBlock{
				Type: "context",
				Elements: []any{
					slack.ImageElement{
						Type:     "image",
						ImageURL: t.User.ProfileImageURL,
						AltText:  t.User.UserName,
					},
					slack.TextObject{
						Type: "plain_text",
						Text: fmt.Sprintf("%s @%s", t.User.Name, t.User.UserName),
					},
				},
			},
			slack.SectionBlock{
				Type: "section",
				Text: slack.TextObject{
					Type:  "plain_text",
					Text:  t.Text,
					Emoji: false,
				},
			},
			slack.SectionBlock{
				Type: "section",
				Text: slack.TextObject{
					Type: "mrkdwn",
					Text: permalink,
				},
			},
		},
	}
}

func main() {
	lambda.Start(DoDisambiguate)
}
