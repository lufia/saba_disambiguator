package main

import (
	"fmt"
	"os"
	"time"

	"encoding/json"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/syou6162/saba_disambiguator/lib"
)

func DoDisambiguate() {
	webhookUrlPositive := os.Getenv("SLACK_WEBHOOK_URL")
	webhookUrlNegative := os.Getenv("SLACK_WEBHOOK_URL_NEGATIVE")

	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	modelJson, err := sabadisambiguator.Asset("model/model.bin")
	if err != nil {
		panic(err)
	}
	model := sabadisambiguator.PerceptronClassifier{}
	err = json.Unmarshal(modelJson, &model)
	if err != nil {
		panic(err)
	}

	search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query:      "mackerel lang:ja exclude:retweets",
		Count:      100,
		ResultType: "recent",
	})

	if err != nil {
		panic(err)
	}

	now := time.Now()

	for _, t := range search.Statuses {
		createdAt, err := t.CreatedAtTime()
		if err != nil {
			panic(err)
		}
		if now.After(createdAt.Add(5 * time.Minute)) {
			continue
		}

		fv := sabadisambiguator.ExtractFeatures(t)
		predLabel := model.Predict(fv)

		tweetPermalink := fmt.Sprintf("https://twitter.com/%s/status/%s", t.User.ScreenName, t.IDStr)
		payload := slack.Payload{
			Text: tweetPermalink,
		}

		if predLabel == sabadisambiguator.POSITIVE {
			fmt.Fprintf(os.Stderr, "%s\n", tweetPermalink)
			err := slack.Send(webhookUrlPositive, "", payload)
			if err != nil {
				panic(err)
			}
		} else if (predLabel == sabadisambiguator.NEGATIVE) && (webhookUrlNegative != "") {
			err := slack.Send(webhookUrlNegative, "", payload)
			if err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	lambda.Start(DoDisambiguate)
}
