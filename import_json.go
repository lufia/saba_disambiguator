package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"encoding/json"
	"strconv"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
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
	cnt := 0
	for stdin.Scan() {
		if err := stdin.Err(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		text := stdin.Text()
		id, err := parseLine(text)
		if err != nil {
			continue
		}

		tweet, resp, err := client.Statuses.Show(id, nil)
		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, resp)
			fmt.Fprintln(os.Stderr, err)
		}

		tweetJson, _ := json.Marshal(tweet)
		fmt.Println(string(tweetJson))
		cnt += 1
		if cnt%10 == 0 {
			time.Sleep(5 * time.Second)
		}
	}
}
