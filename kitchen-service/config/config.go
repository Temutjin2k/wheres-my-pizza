package config

import (
	"kitchen-service/pkg/configparser"
	"kitchen-service/pkg/postgres"
)

type (
	// Config
	Config struct {
		Postgres postgres.Config
	}
)

func New(configpath string) (*Config, error) {
	cfg := &Config{}

	if err := configparser.LoadAndParseYaml(configpath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
