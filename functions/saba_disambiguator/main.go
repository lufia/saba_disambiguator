package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"
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

func DoDisambiguate() error {
	config, err := sabadisambiguator.GetConfigFromFile("config.yml")
	if err != nil {
		return err
	}

	token := oauth1.NewToken(
		config.TwitterConfig.AceessToken,
		config.TwitterConfig.AccessSecret,
	)
	httpClient := oauth1.NewConfig(
		config.TwitterConfig.ConsumerKey,
		config.TwitterConfig.ConsumerSecret,
	).Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

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
			panic(err)
		}
		if now.After(createdAt.Add(5 * time.Minute)) {
			continue
		}

		tweetJson, err := json.Marshal(t)
		if err != nil {
			panic(err)
		}
		fv := sabadisambiguator.ExtractFeatures(t)
		score := model.PredictScore(fv)
		predLabel := model.Predict(fv)
		item := ItemForBigQuery{
			CreatedAt:  createdAt,
			IdStr:      t.IDStr,
			ScreenName: t.User.ScreenName,
			Text:       t.FullText,
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
			err := slack.Send(config.SlackConfig.WebhookUrlPositive, "", payload)
			if err != nil {
				panic(err)
			}
		} else if (predLabel == sabadisambiguator.NEGATIVE) && (config.SlackConfig.WebhookUrlNegative != "") {
			err := slack.Send(config.SlackConfig.WebhookUrlNegative, "", payload)
			if err != nil {
				panic(err)
			}
		}
	}

	if config.BigQueryConfig.ProjectId != "" {
		ctx := context.Background()
		bqClient, err := bigquery.NewClient(ctx, config.BigQueryConfig.ProjectId)
		if err != nil {
			panic(err)
		}
		defer bqClient.Close()

		u := bqClient.Dataset(config.BigQueryConfig.Dataset).Table(config.BigQueryConfig.Table).Inserter()
		err = u.Put(ctx, itemsForBq)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func main() {
	lambda.Start(DoDisambiguate)
}
