package twitter2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Tweet struct {
	ID                string
	Text              string
	CreatedAt         time.Time
	User              *User
	Lang              string
	QuotedStatus      *Tweet
	InReplyToUserName string
	Entities          []Entities
}

type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	UserName        string `json:"username"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
}

type Entities struct {
	// 使われてない
	// Annotations []any `json:"annotations"`
	// 使われていない
	// Cashtags []any `json:"cashtags"`

	Hashtags []EntitiesHashTag `json:"hashtags"`
	Mentions []EntitiesMention `json:"mentions"`
	URLs     []EntitiesURL     `json:"urls"`
}

type EntitiesHashTag struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Tag   string `json:"tag"`
}

type EntitiesMention struct {
	Start    int    `json:"start"`
	End      int    `json:"end"`
	UserName string `json:"username"`
}

type EntitiesURL struct {
	Start       int    `json:"start"`
	End         int    `json:"end"`
	URL         string `json:"url"`
	ExpandedURL string `json:"expaneded_url"`
	DisplayURL  string `json:"display_url"`
	UnwoundURL  string `json:"unwound_url"`
}

type responseRecentSearch struct {
	Data     []responseTweet `json:"data"`
	Includes struct {
		Users  []*User          `json:"users"`
		Tweets []*responseTweet `json:"tweets"`
	} `json:"includes"`
}

type responseTweet struct {
	ID               string            `json:"id"`
	Text             string            `json:"text"`
	CreatedAt        time.Time         `json:"created_at"`
	AuthorID         string            `json:"author_id"`
	Lang             string            `json:"lang"`
	Entities         Entities          `json:"entites"`
	ReferencedTweets []referencedTweet `json:"referenced_tweets"`
	InReplyToUserID  string            `json:"in_reply_to_user_id"`
}

type referencedTweet struct {
	Type referencedTweetType `json:"type"`
	ID   string              `json:"id"`
}

type referencedTweetType string

const (
	typeRetweeted referencedTweetType = "retweeted"
	typeQuoted    referencedTweetType = "quoted"
	typeRepliedTo referencedTweetType = "replied_to"
)

type BearerToken string

var Host = "api.twitter.com"

const RecentSearchPath = "/2/tweets/search/recent"

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent
func (bt BearerToken) RecentSearch(query string) ([]*Tweet, error) {
	params := newParams()
	params.Set("query", query)
	params.Set("max_results", "100")
	params.Set("tweet.fields", "created_at", "entities", "lang", "referenced_tweets")
	params.Set("user.fields", "description")
	params.Set("expansions", "author_id", "in_reply_to_user_id")

	// A space should be escaped into '%20' instead of '+' on twitter's query parameter.
	queryParam := strings.Replace(params.Encode(), "+", "%20", -1)

	u := &url.URL{
		Scheme:   "https",
		Host:     Host,
		Path:     RecentSearchPath,
		RawQuery: queryParam,
	}

	var resp responseRecentSearch
	err := bt.getJSON(&resp, u)
	if err != nil {
		return nil, fmt.Errorf("twitter.RecentSearch: %w", err)
	}

	users := make(map[string]*User, len(resp.Includes.Users))
	for _, u := range resp.Includes.Users {
		users[u.ID] = u
	}
	includesTweets := make(map[string]*responseTweet, len(resp.Includes.Tweets))
	for _, t := range resp.Includes.Tweets {
		includesTweets[t.ID] = t
	}

	tweets := make([]*Tweet, 0, len(resp.Data))
	for _, d := range resp.Data {
		u, ok := users[d.AuthorID]
		if !ok {
			return nil, fmt.Errorf("twitter.RecentSearch: unkown author_id %s", d.AuthorID)
		}

		var quotedStatus *Tweet
		if len(d.ReferencedTweets) > 0 {
			for _, r := range d.ReferencedTweets {
				if r.Type == typeQuoted {
					tw := includesTweets[r.ID]
					if !ok {
						return nil, fmt.Errorf("twitter.RecentSearch: unkown twitter_id %s", t.ID)
					}
					quotedStatus = &Tweet{
						ID:        tw.ID,
						Text:      tw.Text,
						CreatedAt: tw.CreatedAt,
						User:      users[tw.AuthorID],
						Lang:      tw.Lang,
					}
				}
			}
		}

		t := &Tweet{
			ID:                d.ID,
			Text:              d.Text,
			CreatedAt:         d.CreatedAt,
			User:              u,
			Lang:              d.Lang,
			QuotedStatus:      quotedStatus,
			InReplyToUserName: users[d.InReplyToUserID].UserName,
		}
		tweets = append(tweets, t)
	}
	return tweets, nil
}

func (bt BearerToken) newHeader() http.Header {
	p := http.Header{}
	p.Set("Authorization", fmt.Sprintf("Bearer %s", bt))
	p.Set("User-Agent", "sabadisambiguator")
	return p
}

func (bt BearerToken) getJSON(v any, u *url.URL) error {
	return getJSON(v, u, bt.newHeader())
}
