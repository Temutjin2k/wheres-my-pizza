package config

import "github.com/Temutjin2k/wheres-my-pizza/notification-service/pkg/configparser"

type (
	// Config
	Config struct {
	}
)

func New(configpath string) (*Config, error) {
	cfg := &Config{}

	if err := configparser.LoadAndParseYaml(configpath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
