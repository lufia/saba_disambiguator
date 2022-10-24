//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"
)

func parseLine(line string) (int64, error) {
	tokens := strings.Split(line, "/")
	id := tokens[len(tokens)-1]
	return strconv.ParseInt(id, 10, 64)
}

func cacheIdsFromFile(filename string) (map[int64]struct{}, error) {
	cachedIds := make(map[int64]struct{})

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

func main() {
	log.SetFlags(0)
	config, err := sabadisambiguator.GetConfigFromFile("functions/saba_disambiguator/build/config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}
	svc := ssm.New(session.New(), &aws.Config{
		Region: aws.String(config.Region),
	})

	client, err := sabadisambiguator.GetTwitterClient(svc, *config)
	if err != nil {
		log.Fatalf("failed to get Twitter client: %v\n", err)
	}

	cachedIds, err := cacheIdsFromFile(os.Args[1])
	if err != nil {
		log.Fatalf("failed to read cache: %v\n", err)
	}

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
		if _, ok := cachedIds[id]; ok {
			continue
		}

		time.Sleep(1 * time.Second)
		tweet, resp, err := client.Statuses.Show(id, nil)
		if resp.StatusCode != 200 {
			log.Printf("failed to get tweet %s: status=%d, %v\n", text, resp.StatusCode, err)
			continue
		}

		tweetJson, _ := json.Marshal(tweet)
		fmt.Println(string(tweetJson))
	}
}
