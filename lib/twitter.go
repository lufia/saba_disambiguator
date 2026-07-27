package sabadisambiguator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	twitter2 "github.com/syou6162/saba_disambiguator/twitter"
)

type twitterConfig struct {
	BearerToken string
}

func getTwitterConfig(ctx context.Context, svc *ssm.Client, config Config) (twitterConfig, error) {
	twitterConfig := twitterConfig{}

	bearerToken, err := GetValueFromParameterStore(ctx, svc, config.TwitterConfig.ParameterStoreNameBearerToken)
	if err != nil {
		return twitterConfig, err
	}
	twitterConfig.BearerToken = bearerToken
	return twitterConfig, nil
}

func GetTwitterClient(ctx context.Context, svc *ssm.Client, config Config) (*twitter2.Client, error) {
	twitterConfig, err := getTwitterConfig(ctx, svc, config)
	if err != nil {
		return nil, err
	}

	client := twitter2.NewClient(twitterConfig.BearerToken)
	return client, nil
}
