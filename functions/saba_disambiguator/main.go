package main

import (
	"fmt"
	"os"
	"time"

	"encoding/json"

	"github.com/apex/go-apex"
	"github.com/bluele/slack"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/syou6162/saba_disambiguator/lib"
)

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		slackToken := os.Getenv("SLACK_TOKEN")
		channelNamePositive := os.Getenv("SLACK_CHANNEL_NAME")
		channelNameNegative := os.Getenv("SLACK_CHANNEL_NAME_NEGATIVE")

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

		api := slack.New(slackToken)

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
			if predLabel == sabadisambiguator.POSITIVE {
				err := api.ChatPostMessage(channelNamePositive, fmt.Sprintf("https://twitter.com/%s/status/%s", t.User.ScreenName, t.IDStr), nil)
				if err != nil {
					panic(err)
				}
				fmt.Fprintf(os.Stderr, "https://twitter.com/%s/status/%s\n", t.User.ScreenName, t.IDStr)
				// fmt.Fprint(os.Stderr, "%s\n", t.Text)
			} else if (predLabel == sabadisambiguator.NEGATIVE) && (channelNameNegative != "") {
				err := api.ChatPostMessage(channelNameNegative, fmt.Sprintf("https://twitter.com/%s/status/%s", t.User.ScreenName, t.IDStr), nil)
				if err != nil {
					panic(err)
				}
			}
		}
		return nil, nil
	})
}
