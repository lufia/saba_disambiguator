package twitter2

import (
	"fmt"
	"net/http"
	"net/url"
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

const RecentSearchPath = "/2/tweets/search/recent"

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent
func (bt BearerToken) RecentSearch(query string) ([]*Tweet, error) {
	params := newParams()
	params.Set("query", query)
	params.Set("max_results", "100")
	// 検索結果の tweet のオブジェクトに置かれるフィールドを指定する。
	// entities は hashtag, url などの情報
	// referenced_tweets は 検索結果の tweet が言及している tweet (retweeted, quoted など) を含める。
	params.Set("tweet.fields", "created_at", "entities", "lang", "referenced_tweets")
	//　.includes.user にあるユーザー情報に含める要素
	params.Set("user.fields", "description", "id", "name", "username", "url", "profile_image_url")
	// .includes になんの情報を含めるか定める。たとえば author_id を含めると検索結果 tweet の主を .include.user に含める。
	params.Set("expansions", "author_id", "in_reply_to_user_id", "referenced_tweets.id")

	u := &url.URL{
		Scheme:   "https",
		Host:     Host,
		Path:     RecentSearchPath,
		RawQuery: params.Encode(),
	}

	var resp recentSearchResponse
	err := bt.getJSON(&resp, u)
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
				if r.Type == typeQuoted {
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

func (bt BearerToken) newHeader() http.Header {
	p := http.Header{}
	p.Set("Authorization", fmt.Sprintf("Bearer %s", bt))
	p.Set("User-Agent", "sabadisambiguator")
	return p
}

func (bt BearerToken) getJSON(v any, u *url.URL) error {
	return getJSON(v, u, bt.newHeader())
}
