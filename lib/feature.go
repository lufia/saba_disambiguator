package sabadisambiguator

import (
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/ikawaha/kagome.ipadic/tokenizer"
)

type FeatureVector []string

func extractJpnNounFeatures(s string, prefix string) FeatureVector {
	var fv FeatureVector
	if s == "" {
		return fv
	}
	t := tokenizer.New()
	tokens := t.Tokenize(strings.ToLower(s))
	for _, token := range tokens {
		if token.Pos() == "名詞" {
			surface := token.Surface
			if len(token.Features()) >= 2 && token.Features()[1] == "数" {
				surface = "NUM"
			}
			fv = append(fv, prefix+":"+surface)
		}
	}
	return fv
}

func ExtractNounFeatures(s string, prefix string) FeatureVector {
	return extractJpnNounFeatures(s, prefix)
}

func ExtractFeatures(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	text := t.Text
	fv = append(fv, "BIAS")
	fv = append(fv, ExtractNounFeatures(text, "BODY")...)
	return fv
}
