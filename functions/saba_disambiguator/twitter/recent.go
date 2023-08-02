package twitter2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Tweet struct {
	ID        string
	Text      string
	CreatedAt time.Time
	User      *User
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	UserName string `json:"username"`
}

type responseRecentSearch struct {
	Data     []responseTweet `json:"data"`
	Includes struct {
		Users []User `json:"users"`
	} `json:"includes"`
}

type responseTweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	AuthorID  string    `json:"author_id"`
}

type BearerToken string

var Host = "api.twitter.com"

const RecentSearchPath = "/2/tweets/search/recent"

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent
func (bt BearerToken) RecentSearch(query string) ([]*Tweet, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("max_results", "100")
	params.Set("tweet.fields", "created_at")
	params.Set("expansions", "author_id")

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

	users := make(map[string]User, len(resp.Includes.Users))
	for _, u := range resp.Includes.Users {
		users[u.ID] = u
	}

	tweets := make([]*Tweet, 0, len(resp.Data))
	for _, d := range resp.Data {
		u, ok := users[d.AuthorID]
		if !ok {
			return nil, fmt.Errorf("twitter.RecentSearch: unkown author_id %s", d.AuthorID)
		}
		t := &Tweet{
			ID:        d.ID,
			Text:      d.Text,
			CreatedAt: d.CreatedAt,
			User:      &u,
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
