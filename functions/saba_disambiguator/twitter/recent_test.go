package twitter2

import (
	"os"
	"testing"
)

func TestRecentTweet(t *testing.T) {
	c := NewClient(os.Getenv("BEARER_TOKEN"))
	tweets, err := c.RecentSearch("mackerel lang:ja -is:retweet")
	for _, tweet := range tweets {
		t.Logf("%+v %+v", *tweet, *tweet.User)
	}
	if err != nil {
		t.Fatal(err)
	}
}
