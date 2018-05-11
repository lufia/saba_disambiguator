// +build ignore

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
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

type Environment struct {
	TwitterConsumerKey    string `json:"TWITTER_CONSUMER_KEY"`
	TwitterConsumerSecret string `json:"TWITTER_CONSUMER_SECRET"`
	TwitterAccessToken    string `json:"TWITTER_ACCESS_TOKEN"`
	TwitterAccessSecret   string `json:"TWITTER_ACCESS_SECRET"`
}

type ProjectSetting struct {
	Environment Environment `json:environment`
}

func readProjectFile(fileName string) (*ProjectSetting, error) {
	var projectSetting ProjectSetting
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &projectSetting); err != nil {
		return nil, err
	}
	return &projectSetting, nil
}

func main() {
	p, err := readProjectFile("project.json")
	if err != nil {
		panic(err)
	}
	consumerKey := p.Environment.TwitterConsumerKey
	consumerSecret := p.Environment.TwitterConsumerSecret
	accessToken := p.Environment.TwitterAccessToken
	accessSecret := p.Environment.TwitterAccessSecret

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		time.Sleep(1 * time.Second)
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
			continue
		}

		tweetJson, _ := json.Marshal(tweet)
		fmt.Println(string(tweetJson))
	}
}
