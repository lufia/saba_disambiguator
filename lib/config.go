package sabadisambiguator

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type TwitterConfig struct {
	ParameterStoreNameConsumerKey    string `yaml:"parameterStoreNameConsumerKey"`
	ParameterStoreNameConsumerSecret string `yaml:"parameterStoreNameConsumerSecret"`
	ParameterStoreNameAccessToken    string `yaml:"parameterStoreNameAccessToken"`
	ParameterStoreNameAccessSecret   string `yaml:"parameterStoreNameAccessSecret"`
}

type SlackConfig struct {
	ParameterStoreNameWebhookUrlPositive string `yaml:"parameterStoreNameWebhookUrlPositive"`
	ParameterStoreNameWebhookUrlNegative string `yaml:"parameterStoreNameWebhookUrlNegative"`
}

type BigQueryConfig struct {
	ParameterStoreNameServiceAccountCredential string `yaml:"parameterStoreNameServiceAccountCredential"`
	ProjectId                                  string `yaml:"projectId"`
	Dataset                                    string `yaml:"dataset"`
	Table                                      string `yaml:"table"`
}

type Config struct {
	TwitterConfig  TwitterConfig  `yaml:"twitter"`
	SlackConfig    SlackConfig    `yaml:"slack"`
	BigQueryConfig BigQueryConfig `yaml:"bigquery"`
	Query          string         `yaml:"query"`
	Region         string         `yaml:"region"`
}

func GetConfigFromFile(configPath string) (*Config, error) {
	config := Config{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
