package sabadisambiguator

import (
	"math/rand"

	twitter2 "github.com/syou6162/saba_disambiguator/twitter"
)

type LabelType int

const (
	POSITIVE LabelType = 1
	NEGATIVE LabelType = -1
)

type Example struct {
	Label LabelType `json:"Label"`
	Fv    FeatureVector
	Tweet *twitter2.Tweet
}

type Examples []*Example

func NewExampleWithOptions(tweet *twitter2.Tweet, label LabelType, opts ExtractOptions) *Example {
	fv := ExtractFeaturesWithOptions(tweet, opts)
	return &Example{Label: label, Fv: fv, Tweet: tweet}
}

func shuffle(examples Examples) {
	n := len(examples)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		examples[i], examples[j] = examples[j], examples[i]
	}
}
