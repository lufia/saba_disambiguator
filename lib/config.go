package sabadisambiguator

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type TwitterConfig struct {
	ConsumerKey    string `yaml:"consumerKey"`
	ConsumerSecret string `yaml:"consumerSecret"`
	AceessToken    string `yaml:"aceessToken"`
	AccessSecret   string `yaml:"accessSecret"`
}

type SlackConfig struct {
	WebhookUrlPositive string `yaml:"webhookUrlPositive"`
	WebhookUrlNegative string `yaml:"webhookUrlNegative"`
}

type BigQueryConfig struct {
	ProjectId string `yaml:"projectId"`
	Dataset   string `yaml:"dataset"`
	Table     string `yaml:"table"`
}

type Config struct {
	TwitterConfig  TwitterConfig  `yaml:"twitter"`
	SlackConfig    SlackConfig    `yaml:"slack"`
	BigQueryConfig BigQueryConfig `yaml:"bigquery"`
	Query          string         `yaml:"query"`
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
