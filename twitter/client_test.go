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

func TestTweets(t *testing.T) {
	if !*doConnectionTest {
		t.Skip("-conn が指定されていないので skip")
	}
	c := NewClient(os.Getenv("BEARER_TOKEN"))
	tweets, err := c.Tweets([]string{
		"1701003550158200914", // test for a tweet with a quoted tweet
		"1701392004985336253", // for a reply tweet
		"1700576982176850395",
		"1700347488102682862",
		"1700935680795214116", // for a hashtaging
	})
	for _, tweet := range tweets {
		t.Logf("%+v %+v", *tweet, *tweet.User)
	}
	if err != nil {
		t.Fatal(err)
	}
}
