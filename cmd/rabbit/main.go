package main

import (
	"context"
	"log"

	"github.com/Temutjin2k/wheres-my-pizza/config"
	writer "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
)

func main() {
	ctx := context.Background()

	// Init config
	cfg, err := config.New("config.yaml")
	if err != nil {
		log.Fatal("failed to init config", err)
	}

	config.PrintConfig(cfg)

	writer, err := writer.NewClient(ctx, cfg.RabbitMQ)
	if err != nil {
		log.Fatal("failed to create")
	}
	if err := writer.PublishCreateOrder(ctx, &models.CreateOrder{
		Number:          "ORD_20241216_001",
		CustomerName:    "John Doe",
		Type:            types.OrderTypeDelivery,
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

	log.Println("PUBLISHED ORDER")
}

func ptrString(s string) *string {
	return &s
}
