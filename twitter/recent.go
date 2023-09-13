package twitter2

import (
	"fmt"
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

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent
func (c *Client) RecentSearch(q string) ([]*Tweet, error) {
	params := recentSearchQueryParam{
		Query:      q,
		MaxResults: 100,

		// entities は hashtag, url などの情報
		// referenced_tweets は 検索結果の tweet が言及している tweet (retweeted, quoted など) を含める。
		TweetFields: []string{"created_at", "entities", "lang", "referenced_tweets", "author_id"},
		UserFields:  []string{"description", "id", "name", "username", "url", "profile_image_url"},

		// author_id を含めると検索結果 tweet の主を .include.user に含める。
		Expansions: []string{"author_id", "in_reply_to_user_id", "referenced_tweets.id.author_id"},
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
		t, err := tweetResponseToTweet(&d, users, includesTweets)
		if err != nil {
			continue
		}
		tweets = append(tweets, t)
	}
	return tweets, nil
}
