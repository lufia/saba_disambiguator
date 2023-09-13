//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"
	twitter2 "github.com/syou6162/saba_disambiguator/twitter"
)

func parseLine(line string) (string, error) {
	u, err := url.Parse(line)
	if err != nil {
		return "", err
	}
	id := path.Base(u.Path)
	return id, nil
}

func cacheIdsFromFile(filename string) (map[string]struct{}, error) {
	cachedIds := make(map[string]struct{})

	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		text := scanner.Text()
		id, err := parseLine(text)
		if err != nil {
			continue
		}
		cachedIds[id] = struct{}{}
	}
	return cachedIds, nil
}

var flagAppend = flag.String("a", "", "append new tweets to `file`")

type WriteSyncer interface {
	io.Writer
	Sync() error
}

type nopWriter struct{}

func (*nopWriter) Write(p []byte) (int, error) { return len(p), nil }
func (*nopWriter) Sync() error                 { return nil }

func main() {
	log.SetFlags(0)
	flag.Parse()

	config, err := sabadisambiguator.GetConfigFromFile("functions/saba_disambiguator/build/config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		log.Fatalf("failed to get session: %v\n", err)
	}
	svc := ssm.New(sess, &aws.Config{
		Region: aws.String(config.Region),
	})

	client, err := sabadisambiguator.GetTwitterClient(svc, *config)
	if err != nil {
		log.Fatalf("failed to get Twitter client: %v\n", err)
	}

	cachedIds, err := cacheIdsFromFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to read cache: %v\n", err)
	}

	var w WriteSyncer = &nopWriter{}
	if *flagAppend != "" {
		f, err := os.OpenFile(*flagAppend, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("failed to open '%s': %v\n", *flagAppend, err)
		}
		defer f.Close()
		w = f
	}

	newTweetsID := []string{}
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		text := stdin.Text()
		id, err := parseLine(text)
		if err != nil {
			continue
		}
		if _, ok := cachedIds[id]; ok {
			continue
		}
		newTweetsID = append(newTweetsID, id)
	}

	tweets, err := fetchTweets(client, newTweetsID)
	if err != nil {
		log.Fatalf("failed to get tweet: %v\n", err)
	}

	for _, tweet := range tweets {
		tweetJson, _ := json.Marshal(tweet)
		fmt.Println(string(tweetJson))
		fmt.Fprintln(w, string(tweetJson))
		if err := stdin.Err(); err != nil {
			log.Fatalln(err)
		}
	}

	if err := w.Sync(); err != nil {
		log.Fatalf("failed to flush tweets: %v\n", err)
	}
}

func fetchTweets(client *twitter2.Client, ids []string) ([]*twitter2.Tweet, error) {
	// the `tweets` API accepts less than 100 tweets.
	const N = 100
	c := 0
	var tweets []*twitter2.Tweet
	for start := 0; start < len(ids); start += N {
		end := start + N
		if end > len(ids) {
			end = len(ids)
		}
		t, err := client.Tweets(ids[start:end])
		if err != nil {
			return nil, err
		}
		c += len(t)
		log.Printf("fetched %d tweets...", c)
		tweets = append(tweets, t...)
		// avoid `too many requests`
		time.Sleep(time.Minute)
	}
	return tweets, nil
}
