package sabadisambiguator

import (
	"strings"

	"strconv"

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

func inReplyToScreenName(t twitter.Tweet) string {
	return t.InReplyToScreenName
}

func containsMackerelInScreenName(screenName string) bool {
	return strings.Contains(strings.ToLower(screenName), "mackerel")
}

func includeMackerelInUserMentions(t twitter.Tweet) bool {
	result := false
	for _, m := range t.Entities.UserMentions {
		if containsMackerelInScreenName(m.ScreenName) {
			return true
		}
	}
	return result
}

func includeMackerelInReplyToScreenName(t twitter.Tweet) bool {
	return containsMackerelInScreenName(t.InReplyToScreenName)
}

func lang(t twitter.Tweet) string {
	return t.Lang
}

func ExtractNounFeatures(s string, prefix string) FeatureVector {
	return extractJpnNounFeatures(s, prefix)
}

func extractNounFeaturesFromQuotedText(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	if t.QuotedStatus == nil {
		return fv
	}
	return ExtractNounFeatures(t.QuotedStatus.Text, "QuotedText")
}

func extractNounFeaturesFromUserDescription(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	if t.QuotedStatus == nil {
		return fv
	}
	return ExtractNounFeatures(t.User.Description, "UserDEscription")
}

func screenNameInQuotedStatus(t twitter.Tweet) string {
	result := ""
	if t.QuotedStatus == nil {
		return result
	}
	return t.QuotedStatus.User.ScreenName
}

func domainsInEntities(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	if t.Entities == nil {
		return fv
	}

	for _, u := range t.Entities.Urls {
		fv = append(fv, "DomainsInEntity:"+strings.Split(u.DisplayURL, "/")[0])
	}
	return fv
}

func hashtagsInEntities(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	if t.Entities == nil {
		return fv
	}

	for _, h := range t.Entities.Hashtags {
		fv = append(fv, "HashtagsInEntity:"+h.Text)
	}
	return fv
}

func ExtractFeatures(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	text := t.Text

	fv = append(fv, "BIAS")
	fv = append(fv, "ScreenName:"+t.User.ScreenName)
	fv = append(fv, "inReplyToScreenName:"+inReplyToScreenName(t))
	fv = append(fv, "screenNameInQuotedStatus"+screenNameInQuotedStatus(t))
	fv = append(fv, "lang:"+lang(t))
	fv = append(fv, "containsMackerelInScreenName:"+strconv.FormatBool(containsMackerelInScreenName(t.User.ScreenName)))
	fv = append(fv, "includeMackerelInUserMentions:"+strconv.FormatBool(includeMackerelInUserMentions(t)))
	fv = append(fv, "includeMackerelInReplyToScreenName:"+strconv.FormatBool(includeMackerelInReplyToScreenName(t)))

	fv = append(fv, ExtractNounFeatures(text, "Text")...)
	fv = append(fv, extractNounFeaturesFromQuotedText(t)...)
	fv = append(fv, extractNounFeaturesFromUserDescription(t)...)
	fv = append(fv, domainsInEntities(t)...)
	fv = append(fv, hashtagsInEntities(t)...)
	return fv
}
