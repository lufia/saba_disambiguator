package sabadisambiguator

import (
	"github.com/dghubble/go-twitter/twitter"
	"math/rand"
)

type LabelType int

const (
	POSITIVE  LabelType = 1
	NEGATIVE  LabelType = -1
)

type Example struct {
	Label       LabelType `json:"Label"`
	Fv          FeatureVector
	Tweet twitter.Tweet
}

type Examples []*Example

func NewExample(tweet twitter.Tweet, label LabelType) *Example {
	fv := ExtractFeatures(tweet)
	return &Example{Label:label, Fv: fv, Tweet: tweet}
}

func shuffle(examples Examples) {
	n := len(examples)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		examples[i], examples[j] = examples[j], examples[i]
	}
}

