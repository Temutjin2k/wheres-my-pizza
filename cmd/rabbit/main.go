package main

import (
	"context"
	"log"
	"time"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	writer "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

func main() {
	ctx := context.Background()

	// Init config
	cfg, err := config.New("config.yaml")
	if err != nil {
		log.Fatal("failed to init config", err)
	}

	config.PrintConfig(cfg)

	logger := logger.InitLogger("rabbit_producer", logger.LevelDebug)

	writer, err := writer.NewOrderProducer(ctx, cfg.RabbitMQ, logger)
	if err != nil {
		log.Fatal("failed to create")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*7)
	defer cancel()
	if err := writer.PublishCreateOrder(ctx, &models.CreateOrder{
		Number:          "ORD_20241216_001",
		CustomerName:    "John Doe",
		Type:            types.OrderTypeTakeOut,
		TableNumber:     nil,
		DeliveryAddress: ptrString("123 Main St, City"),
		Items: []models.CreateOrderItem{
			{
				Name:     "Margherita Pizza",
				Quantity: 7,
				Price:    15.99,
			},
		},
		TotalAmount: 111.93,
		Priority:    10,
	}); err != nil {
		log.Fatal(err)
	}

	logger.Info(ctx, types.ActionOrderPublished, "PUBLISHED")
}

func ptrString(s string) *string {
	return &s
}
