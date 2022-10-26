package sabadisambiguator

import (
	"net/url"
	"strconv"
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

func inReplyToScreenName(t twitter.Tweet) string {
	return t.InReplyToScreenName
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
	return ExtractNounFeatures(t.User.Description, "UserDescription")
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

func wordsInUrlPaths(t twitter.Tweet) FeatureVector {
	var fv FeatureVector
	if t.Entities == nil {
		return fv
	}

	for _, url_ := range t.Entities.Urls {
		u, err := url.Parse(url_.ExpandedURL)
		if err != nil {
			continue
		}
		for _, w := range strings.Split(u.Path, "/") {
			if w == "" {
				continue
			}
			fv = append(fv, "wordsInUrlPaths:"+w)
		}
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

type ExtractOptions struct {
	ScreenNames []string
}

func (opts *ExtractOptions) contains(screenName string) bool {
	screenName = strings.ToLower(screenName)

	// for backward compatibility
	if len(opts.ScreenNames) == 0 {
		return strings.Contains(screenName, "mackerel")
	}

	for _, s := range opts.ScreenNames {
		if s == screenName {
			return true
		}
	}
	return false
}

func (opts *ExtractOptions) includeScreenNameInUserMentions(t twitter.Tweet) bool {
	result := false
	for _, m := range t.Entities.UserMentions {
		if opts.contains(m.ScreenName) {
			return true
		}
	}
	return result
}

func (opts *ExtractOptions) includeScreenNameInReplyToScreenName(t twitter.Tweet) bool {
	return opts.contains(t.InReplyToScreenName)
}

func ExtractFeatures(t twitter.Tweet) FeatureVector {
	return ExtractFeaturesWithOptions(t, ExtractOptions{})
}

func ExtractFeaturesWithOptions(t twitter.Tweet, opts ExtractOptions) FeatureVector {
	var fv FeatureVector
	text := t.Text

	fv = append(fv, "BIAS")
	fv = append(fv, "ScreenName:"+t.User.ScreenName)
	fv = append(fv, "inReplyToScreenName:"+inReplyToScreenName(t))
	fv = append(fv, "screenNameInQuotedStatus:"+screenNameInQuotedStatus(t))
	fv = append(fv, "lang:"+lang(t))
	fv = append(fv, "containsMackerelInScreenName:"+strconv.FormatBool(opts.contains(t.User.ScreenName)))
	fv = append(fv, "includeMackerelInUserMentions:"+strconv.FormatBool(opts.includeScreenNameInUserMentions(t)))
	fv = append(fv, "includeMackerelInReplyToScreenName:"+strconv.FormatBool(opts.includeScreenNameInReplyToScreenName(t)))

	fv = append(fv, ExtractNounFeatures(text, "Text")...)
	fv = append(fv, extractNounFeaturesFromQuotedText(t)...)
	fv = append(fv, extractNounFeaturesFromUserDescription(t)...)
	fv = append(fv, domainsInEntities(t)...)
	fv = append(fv, hashtagsInEntities(t)...)
	fv = append(fv, wordsInUrlPaths(t)...)
	return fv
}
