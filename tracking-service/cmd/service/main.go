package main

import (
	"context"

	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/config"
	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/internal/app"
	"github.com/Temutjin2k/wheres-my-pizza/tracking-service/pkg/logger"
)

const (
	serviceName = "tracking-service"
	configPath  = "config.yaml"
)

func main() {
	ctx := context.Background()

	// Init logger
	log := logger.InitLogger(serviceName, logger.LevelDebug)

	// Init config
	cfg, err := config.New(configPath)
	if err != nil {
		log.Error(ctx, "config_init", "failed to init config", err)
		return
	}

	config.PrintConfig(cfg)

	// Creating application
	app, err := app.NewApplication(ctx, cfg, log)
	if err != nil {
		log.Error(ctx, "app_init", "failed to init application", err)
		return
	}

	// Running the apllication
	err = app.Run()
	if err != nil {
		log.Error(ctx, "app_run", "failed to run application", err)
		return
	}
}
