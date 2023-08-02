package twitter2

import "time"

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

	Hashtags []HashtagEntry `json:"hashtags"`
	Mentions []MentionEntry `json:"mentions"`
	URLs     []URLEntry     `json:"urls"`
}

type HashtagEntry struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Tag   string `json:"tag"`
}

type MentionEntry struct {
	Start    int    `json:"start"`
	End      int    `json:"end"`
	UserName string `json:"username"`
}

type URLEntry struct {
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
