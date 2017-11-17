package main

import (
	"fmt"
	"os"

	"encoding/json"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/syou6162/saba_disambiguator/lib"
)

func main() {
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
		// Query:      "mackerel lang:ja until:2017-11-16",
		Query:      "mackerel lang:ja",
		Count:      100,
		ResultType: "recent",
	})

	if err != nil {
		panic(err)
	}

	for _, t := range search.Statuses {
		fv := sabadisambiguator.ExtractFeatures(t)
		if model.Predict(fv) == sabadisambiguator.POSITIVE {
			fmt.Fprintf(os.Stderr, "https://twitter.com/%s/status/%s\n", t.User.ScreenName, t.IDStr)
			fmt.Fprint(os.Stderr, "%s\n", t.Text)
		}
	}
}
