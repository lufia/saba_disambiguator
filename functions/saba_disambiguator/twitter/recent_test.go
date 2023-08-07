package twitter2

import (
	"flag"
	"os"
	"testing"
)

var doConnectionTest = flag.Bool("conn", false, "twitter に接続したテストをする")

func TestRecentTweet(t *testing.T) {
	if !*doConnectionTest {
		t.Skip("-conn が指定されていないので skip")
	}
	c := NewClient(os.Getenv("BEARER_TOKEN"))
	tweets, err := c.RecentSearch("mackerel lang:ja -is:retweet")
	for _, tweet := range tweets {
		t.Logf("%+v %+v", *tweet, *tweet.User)
	}
	if err != nil {
		t.Fatal(err)
	}
}
