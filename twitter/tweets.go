package twitter2

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

type tweetsByIDQueryParam struct {
	Expansions  []string `url:"expansions,comma"`
	TweetFields []string `url:"tweet.fields,comma"`
	UserFields  []string `url:"user.fields,comma"`
}

type tweetsResponse struct {
	// 検索によって返される tweet
	Data tweetResponse `json:"data"`

	// twitter v2 の　API は ユーザー や tweet に関係する他の tweet を ID で表記されている。
	// それらの実態は includes 以下におかれる。 (query parameter に includes に追加するように指定する必要がある。)
	Includes struct {
		Users  []*User          `json:"users"`
		Tweets []*tweetResponse `json:"tweets"`
	} `json:"includes"`
}

const TweetsPath = "/2/tweets/"

func (c *Client) TweetsByID(id string) (*Tweet, error) {
	params := &tweetsByIDQueryParam{
		// entities は hashtag, url などの情報
		// referenced_tweets は 検索結果の tweet が言及している tweet (retweeted, quoted など) を含める。
		TweetFields: []string{"created_at", "entities", "lang", "referenced_tweets", "author_id"},
		UserFields:  []string{"description", "id", "name", "username", "url", "profile_image_url"},

		// author_id を含めると検索結果 tweet の主を .include.user に含める。
		Expansions: []string{"author_id", "in_reply_to_user_id", "referenced_tweets.id"},
	}
	v, err := query.Values(params)
	if err != nil {
		return nil, fmt.Errorf("twitter.tweets: %w", err)
	}
	path, err := url.JoinPath(TweetsPath, id)
	if err != nil {
		return nil, fmt.Errorf("twitter.tweets: %w", err)
	}
	u := &url.URL{
		Scheme: "https",
		Host:   Host,
		Path:   path,
		// A space should be escaped into '%20' instead of '+' on twitter's query parameter.
		RawQuery: strings.Replace(v.Encode(), "+", "%20", -1),
	}

	var resp tweetsResponse
	err = c.getJSON(&resp, u)
	if err != nil {
		return nil, fmt.Errorf("twitter.tweets: %w", err)
	}
	fmt.Printf("%+v\n", resp)
	users := make(map[string]*User, len(resp.Includes.Users))
	for _, u := range resp.Includes.Users {
		users[u.ID] = u
	}
	includesTweets := make(map[string]*tweetResponse, len(resp.Includes.Tweets))
	for _, t := range resp.Includes.Tweets {
		includesTweets[t.ID] = t
	}
	return tweetResponseToTweet(&resp.Data, users, includesTweets)
}
