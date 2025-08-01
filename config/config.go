package config

import (
	"github.com/Temutjin2k/wheres-my-pizza/pkg/configparser"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/postgres"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/rabbit"
)

type (
	// Config
	Config struct {
		Mode     string
		Server   Server
		Postgres postgres.Config
		RabbitMQ rabbit.Config
	}

	// Servers config
	Server struct {
		HTTPServer HTTPServer
	}

	// HTTP service
	HTTPServer struct {
		Host string `env:"HTTP_HOST" default:"0.0.0.0"`
		Port int    `env:"HTTP_PORT" default:"8080"`
	}
)

func New(filepath string) (*Config, error) {
	cfg := &Config{}

	if err := configparser.LoadAndParseYaml(filepath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
