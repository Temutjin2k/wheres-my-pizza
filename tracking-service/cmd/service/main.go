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
		log.Error(ctx, "failed to init config", "error", err)
		return
	}

	config.PrintConfig(cfg)

	// Creating application
	app, err := app.NewApplication(ctx, cfg, log)
	if err != nil {
		log.Error(ctx, "failed to init application", "error", err)
		return
	}

	// Running the apllication
	err = app.Run()
	if err != nil {
		log.Error(ctx, "failed to run application", "error", err)
		return
	}
}
