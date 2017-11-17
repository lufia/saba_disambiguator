package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"encoding/json"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"strconv"
)

func parseLine(line string) (int64, error) {
	tokens := strings.Split(line, "/")
	id := tokens[len(tokens)-1]
	return strconv.ParseInt(id, 10, 64)
}

func main() {
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		if err := stdin.Err(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		text := stdin.Text()
		id, err := parseLine(text)
		if err != nil {
			continue
		}

		tweet, _, _ := client.Statuses.Show(id, nil)
		tweetJson, _ := json.Marshal(tweet)
		fmt.Println(string(tweetJson))
	}
}
