package main

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/syou6162/saba_disambiguator/lib"
)

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

		e := sabadisambiguator.NewExample(t, label)
		examples = append(examples, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return examples, nil
}

func main() {
	examplesPos, err := readExamplesFromFile(os.Args[1], sabadisambiguator.POSITIVE)
	if err != nil {
		panic(err)
	}

	examplesNeg, err := readExamplesFromFile(os.Args[2], sabadisambiguator.NEGATIVE)
	if err != nil {
		panic(err)
	}

	examples := append(examplesPos, examplesNeg...)
	p := sabadisambiguator.NewPerceptronClassifier(examples)
	sabadisambiguator.WritePerceptron(*p, "model/model.bin")
}
