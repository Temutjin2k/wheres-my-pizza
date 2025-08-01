package config

import (
	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/pkg/configparser"
	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/pkg/postgres"
)

type (
	// Config
	Config struct {
		Server   Server
		Postgres postgres.Config
	}

	// Servers config
	Server struct {
		HTTPServer HTTPServer
	}

	// HTTP service
	HTTPServer struct {
		Port int `env:"HTTP_PORT" default:"8080"`
	}
)

func New(filepath string) (*Config, error) {
	cfg := &Config{}

	if err := configparser.LoadAndParseYaml(filepath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
