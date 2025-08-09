package main

import (
	"fmt"
	"strings"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
)

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	"github.com/Temutjin2k/wheres-my-pizza/config"
// 	rab "github.com/Temutjin2k/wheres-my-pizza/internal/adapter/rabbit"
// 	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
// 	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
// 	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
// )

// func main() {
// 	ctx := context.Background()

// 	// Init config
// 	cfg, err := config.New("config.yaml")
// 	if err != nil {
// 		log.Fatal("failed to init config", err)
// 	}

// 	config.PrintConfig(cfg)

// 	logger := logger.InitLogger("rabbit_consumer", logger.LevelDebug)
// 	consumer, err := rab.NewOrderConsumer(ctx, cfg.RabbitMQ, 5, []string{types.OrderTypeDelivery}, logger)
// 	if err != nil {
// 		log.Fatal("failed to create", err)
// 	}

// 	// Will work 30 second
// 	// ctx, cancel := context.WithTimeout(ctx, time.Second*5)
// 	// defer cancel()

// 	err = consumer.Consume(ctx, types.OrderTypeDelivery, hand)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// // func ptrString(s string) *string {
// // 	return &s
// // }

// func hand(ord *models.CreateOrder) error {
// 	fmt.Println("GOT MESSAGE")
// 	fmt.Println(ord)
// 	return nil
// }

func main() {
	fl := "bro,bro,dine_in,delivery"

	arr := strings.Split(fl, ",")
	fmt.Println("ARR: ", arr)
	for _, ot := range arr {
		if !types.IsValidOrderType(ot) {
			fmt.Println("INVALID: ", ot)
		} else {
			fmt.Println("VALID: ", ot)
		}
	}

	fmt.Println(strings.Join(arr, ","))
}
