package twitter2

import (
	"fmt"
	"time"
)

// twitter API v2 が返してくる tweet object は user 情報が id で返されるなど扱いにくい。
// そこで、 id で返される要素を他の要素と join して返している。
type Tweet struct {
	ID                string
	Text              string
	CreatedAt         time.Time
	User              *User
	Lang              string
	QuotedStatus      *Tweet
	InReplyToUserName string
	Entities          Entities
}

// https://developer.twitter.com/en/docs/twitter-api/data-dictionary/object-model/user
type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	UserName        string `json:"username"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent#:~:text=regarding%20referenced%20entity.-,entities,-object
type Entities struct {
	// saba_disambiguator で使われてない
	// Annotations []any `json:"annotations"`
	// 使われていない
	// Cashtags []any `json:"cashtags"`

	Hashtags []HashtagEntry `json:"hashtags"`
	Mentions []MentionEntry `json:"mentions"`
	URLs     []URLEntry     `json:"urls"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent#:~:text=full%20destination%20URL.-,entities.hashtags,-array
type HashtagEntry struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Tag   string `json:"tag"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent#:~:text=of%20the%20Hashtag.-,entities.mentions,-array
type MentionEntry struct {
	Start    int    `json:"start"`
	End      int    `json:"end"`
	UserName string `json:"username"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent#:~:text=the%20annotation%20type.-,entities.urls,-array
type URLEntry struct {
	Start       int    `json:"start"`
	End         int    `json:"end"`
	URL         string `json:"url"`
	ExpandedURL string `json:"expaneded_url"`
	DisplayURL  string `json:"display_url"`
	UnwoundURL  string `json:"unwound_url"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent
type tweetResponse struct {
	ID               string            `json:"id"`
	Text             string            `json:"text"`
	CreatedAt        time.Time         `json:"created_at"`
	AuthorID         string            `json:"author_id"`
	Lang             string            `json:"lang"`
	Entities         Entities          `json:"entities"`
	ReferencedTweets []referencedTweet `json:"referenced_tweets"`
	InReplyToUserID  string            `json:"in_reply_to_user_id"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent#:~:text=request%27s%20query%20parameter.-,referenced_tweets,-array
type referencedTweet struct {
	Type referencedTweetType `json:"type"`
	ID   string              `json:"id"`
}

// https://developer.twitter.com/en/docs/twitter-api/tweets/search/api-reference/get-tweets-search-recent#:~:text=referenced_tweets.type
type referencedTweetType string

const (
	typeRetweeted referencedTweetType = "retweeted"
	typeQuoted    referencedTweetType = "quoted"
	typeRepliedTo referencedTweetType = "replied_to"
)

func tweetResponseToTweet(t *tweetResponse, users map[string]*User, includesTweets map[string]*tweetResponse) (*Tweet, error) {
	// tweet オブジェクトに含まれるのは auther_id (数字の並び)のみ
	// .includes.user から対応するユーザーを持ってくる。
	u, ok := users[t.AuthorID]
	if !ok {
		return nil, fmt.Errorf("twitter.tweetResponseToTweet: unkown author_id %s", t.AuthorID)
	}

	// saba_disambiguator は referenced tweets (retweeted, quoted, replied to) の中で quote された tweet のみ見ている。
	var quotedStatus *Tweet
	if len(t.ReferencedTweets) > 0 {
		for _, r := range t.ReferencedTweets {
			if r.Type != typeQuoted {
				continue
			}
			tw, ok := includesTweets[r.ID]
			if !ok {
				return nil, fmt.Errorf("twitter.tweetResponseToTweet: unkown twitter_id %s", r.ID)
			}
			twu, ok := users[tw.AuthorID]
			if !ok {
				return nil, fmt.Errorf("twitter.tweetResponseToTweet: unkown user_id %s", tw.AuthorID)
			}
			quotedStatus = &Tweet{
				ID:        tw.ID,
				Text:      tw.Text,
				CreatedAt: tw.CreatedAt,
				User:      twu,
				Lang:      tw.Lang,
			}
		}
	}

	inReplyToUserName := ""
	if t.InReplyToUserID != "" {
		u, ok := users[t.InReplyToUserID]
		if !ok {
			return nil, fmt.Errorf("twitter.tweetResponseToTweet: unkown user_id %s", t.InReplyToUserID)
		}
		inReplyToUserName = u.UserName
	}

	return &Tweet{
		ID:                t.ID,
		Text:              t.Text,
		CreatedAt:         t.CreatedAt,
		User:              u,
		Lang:              t.Lang,
		QuotedStatus:      quotedStatus,
		InReplyToUserName: inReplyToUserName,
		Entities:          t.Entities,
	}, nil
}
