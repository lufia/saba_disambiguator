// +build ignore

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
	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"
)

func parseLine(line string) (int64, error) {
	tokens := strings.Split(line, "/")
	id := tokens[len(tokens)-1]
	return strconv.ParseInt(id, 10, 64)
}

func main() {
	config, err := sabadisambiguator.GetConfigFromFile("functions/saba_disambiguator/build/config.yml")
	if err != nil {
		panic(err)
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
