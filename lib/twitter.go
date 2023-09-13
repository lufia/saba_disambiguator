package sabadisambiguator

import (
	"github.com/aws/aws-sdk-go/service/ssm"
	twitter2 "github.com/syou6162/saba_disambiguator/twitter"
)

type twitterConfig struct {
	BearerToken string
}

func getTwitterConfig(svc *ssm.SSM, config Config) (twitterConfig, error) {
	twitterConfig := twitterConfig{}

	bearerToken, err := GetValueFromParameterStore(svc, config.TwitterConfig.ParameterStoreNameBearerToken)
	if err != nil {
		return twitterConfig, err
	}
	twitterConfig.BearerToken = bearerToken
	return twitterConfig, nil
}

func GetTwitterClient(svc *ssm.SSM, config Config) (*twitter2.Client, error) {
	twitterConfig, err := getTwitterConfig(svc, config)
	if err != nil {
		return nil, err
	}

	client := twitter2.NewClient(twitterConfig.BearerToken)
	return client, nil
}
