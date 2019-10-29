package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"
)

func DoDisambiguate() error {
	config, err := sabadisambiguator.GetConfigFromFile("build/config.yml")
	if err != nil {
		return err
	}
	fmt.Println(*config)

	token := oauth1.NewToken(
		config.TwitterConfig.AceessToken,
		config.TwitterConfig.AccessSecret,
	)
	httpClient := oauth1.NewConfig(
		config.TwitterConfig.ConsumerKey,
		config.TwitterConfig.ConsumerSecret,
	).Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	model, err := sabadisambiguator.LoadPerceptron("build/model.bin")
	if err != nil {
		return err
	}

	search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query:      "mackerel lang:ja exclude:retweets",
		Count:      100,
		ResultType: "recent",
	})

	if err != nil {
		return err
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
		fmt.Println(tweetPermalink)

		continue

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
	return nil
}

func main() {
	err := DoDisambiguate()
	if err != nil {
		fmt.Println(err)
	}
	// lambda.Start(DoDisambiguate)
}
