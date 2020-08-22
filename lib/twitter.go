package sabadisambiguator

import (
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type twitterConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func getTwitterConfig(svc *ssm.SSM, config Config) (twitterConfig, error) {
	twitterConfig := twitterConfig{}

	consumerKey, err := GetValueFromParameterStore(svc, config.TwitterConfig.ParameterStoreNameConsumerKey)
	if err != nil {
		return twitterConfig, err
	}
	twitterConfig.ConsumerKey = consumerKey

	consumerSecret, err := GetValueFromParameterStore(svc, config.TwitterConfig.ParameterStoreNameConsumerSecret)
	if err != nil {
		return twitterConfig, err
	}
	twitterConfig.ConsumerSecret = consumerSecret

	accessToken, err := GetValueFromParameterStore(svc, config.TwitterConfig.ParameterStoreNameAccessToken)
	if err != nil {
		return twitterConfig, err
	}
	twitterConfig.AccessToken = accessToken

	accessSecret, err := GetValueFromParameterStore(svc, config.TwitterConfig.ParameterStoreNameAccessSecret)
	if err != nil {
		return twitterConfig, err
	}
	twitterConfig.AccessSecret = accessSecret

	return twitterConfig, nil
}

func GetTwitterClient(svc *ssm.SSM, config Config) (*twitter.Client, error) {
	twitterConfig, err := getTwitterConfig(svc, config)
	if err != nil {
		return nil, err
	}

	token := oauth1.NewToken(twitterConfig.AccessToken, twitterConfig.AccessSecret)
	httpClient := oauth1.NewConfig(
		twitterConfig.ConsumerKey,
		twitterConfig.ConsumerSecret,
	).Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	return client, nil
}
