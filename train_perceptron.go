//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	sabadisambiguator "github.com/syou6162/saba_disambiguator/lib"
)

var config *sabadisambiguator.Config

func parseLine(line string) (twitter.Tweet, error) {
	var tweet twitter.Tweet
	err := json.Unmarshal([]byte(line), &tweet)
	return tweet, err
}

func readExamplesFromFile(fileName string, label sabadisambiguator.LabelType) (sabadisambiguator.Examples, error) {
	var examples sabadisambiguator.Examples
	fp, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		text := scanner.Text()

		t, err := parseLine(text)
		if err != nil || t.ID == 0 {
			continue
		}

		e := sabadisambiguator.NewExampleWithOptions(t, label, sabadisambiguator.ExtractOptions{
			ScreenNames: config.ScreenNames,
		})
		examples = append(examples, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return examples, nil
}

func main() {
	log.SetFlags(0)

	cfg, err := sabadisambiguator.GetConfigFromFile("functions/saba_disambiguator/build/config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err)
	}
	config = cfg

	examplesPos, err := readExamplesFromFile(os.Args[1], sabadisambiguator.POSITIVE)
	if err != nil {
		log.Fatalf("failed to read %s: %v\n", os.Args[1], err)
	}

	examplesNeg, err := readExamplesFromFile(os.Args[2], sabadisambiguator.NEGATIVE)
	if err != nil {
		log.Fatalf("failed to read %s: %v\n", os.Args[2], err)
	}

	examples := append(examplesPos, examplesNeg...)
	p := sabadisambiguator.NewPerceptronClassifier(examples)
	sabadisambiguator.WritePerceptron(*p, "functions/saba_disambiguator/build/model.bin")
}
