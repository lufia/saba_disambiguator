package twitter2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

type BearerToken string

var Host = "api.twitter.com"

type recentSearchResponse struct {
	// 検索によって返される tweet
	Data []tweetResponse `json:"data"`

	// twitter v2 の　API は ユーザー や tweet に関係する他の tweet を ID で表記されている。
	// それらの実態は includes 以下におかれる。 (query parameter に includes に追加するように指定する必要がある。)
	Includes struct {
		Users  []*User          `json:"users"`
		Tweets []*tweetResponse `json:"tweets"`
	} `json:"includes"`
}

type recentSearchQueryParam struct {
	Query      string `url:"query"`
	MaxResults int    `url:"max_results"`

	// 検索結果の tweet のオブジェクトに置かれるフィールドを指定する。
	TweetFields []string `url:"tweet.fields,comma"`

	//　.includes.user にあるユーザー情報に含める要素
	UserFields []string `url:"user.fields,comma"`

	// .includes になんの情報を含めるか定める。
	Expansions []string `url:"expansions,comma"`
}

const RecentSearchPath = "/2/tweets/search/recent"

type Client struct {
	bearerToken string
}

func NewClient(bearerToken string) *Client {
	return &Client{bearerToken: bearerToken}
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent
func (c *Client) RecentSearch(q string) ([]*Tweet, error) {
	params := recentSearchQueryParam{
		Query:      q,
		MaxResults: 100,

		// entities は hashtag, url などの情報
		// referenced_tweets は 検索結果の tweet が言及している tweet (retweeted, quoted など) を含める。
		TweetFields: []string{"created_at", "entities", "lang", "referenced_tweets"},
		UserFields:  []string{"description", "id", "name", "username", "url", "profile_image_url"},

		// author_id を含めると検索結果 tweet の主を .include.user に含める。
		Expansions: []string{"author_id", "in_reply_to_user_id", "referenced_tweets.id"},
	}
	v, err := query.Values(params)
	if err != nil {
		return nil, fmt.Errorf("twitter.RecentSearch: %w", err)
	}

	u := &url.URL{
		Scheme: "https",
		Host:   Host,
		Path:   RecentSearchPath,
		// A space should be escaped into '%20' instead of '+' on twitter's query parameter.
		RawQuery: strings.Replace(v.Encode(), "+", "%20", -1),
	}

	var resp recentSearchResponse
	err = c.getJSON(&resp, u)
	if err != nil {
		return nil, fmt.Errorf("twitter.RecentSearch: %w", err)
	}

	users := make(map[string]*User, len(resp.Includes.Users))
	for _, u := range resp.Includes.Users {
		users[u.ID] = u
	}
	includesTweets := make(map[string]*tweetResponse, len(resp.Includes.Tweets))
	for _, t := range resp.Includes.Tweets {
		includesTweets[t.ID] = t
	}

	tweets := make([]*Tweet, 0, len(resp.Data))
	for _, d := range resp.Data {
		// tweet オブジェクトに含まれるのは auther_id (数字の並び)のみ
		// .includes.user から対応するユーザーを持ってくる。
		u, ok := users[d.AuthorID]
		if !ok {
			return nil, fmt.Errorf("twitter.RecentSearch: unkown author_id %s", d.AuthorID)
		}

		// saba_disambiguator は referenced tweets (retweeted, quoted, replied to) の中で quote された tweet のみ見ている。
		var quotedStatus *Tweet
		if len(d.ReferencedTweets) > 0 {
			for _, r := range d.ReferencedTweets {
				if r.Type != typeQuoted {
					continue
				}
				tw, ok := includesTweets[r.ID]
				if !ok {
					return nil, fmt.Errorf("twitter.RecentSearch: unkown twitter_id %s", r.ID)
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

		inReplyToUserName := ""
		if d.InReplyToUserID != "" {
			u, ok := users[d.InReplyToUserID]
			if !ok {
				return nil, fmt.Errorf("twitter.RecentSearch: unkown user_id %s", d.InReplyToUserID)
			}
			inReplyToUserName = u.UserName
		}

		t := &Tweet{
			ID:                d.ID,
			Text:              d.Text,
			CreatedAt:         d.CreatedAt,
			User:              u,
			Lang:              d.Lang,
			QuotedStatus:      quotedStatus,
			InReplyToUserName: inReplyToUserName,
			Entities:          d.Entities,
		}
		tweets = append(tweets, t)
	}
	return tweets, nil
}

func (c *Client) newHeader() http.Header {
	p := http.Header{}
	p.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	p.Set("User-Agent", "sabadisambiguator")
	return p
}

func (c *Client) getJSON(v any, u *url.URL) error {
	return getJSON(v, u, c.newHeader())
}
