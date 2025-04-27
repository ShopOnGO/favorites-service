package main

import (
	"fmt"

	"github.com/ShopOnGO/favorites-service/configs"
	"github.com/ShopOnGO/favorites-service/migrations"
	"github.com/ShopOnGO/favorites-service/pkg/db"
	"github.com/gin-gonic/gin"
)

func main() {
	migrations.CheckForMigrations()
	conf := configs.LoadConfig()
	// database := db.NewDB(conf)
	_ = db.NewDB(conf)
	router := gin.Default()

	// repository

	// service

	// handler

	// kafkaProductConsumer := kafkaService.NewConsumer(
	// 	conf.KafkaProduct.Brokers,
	// 	conf.KafkaProduct.Topic,
	// 	conf.KafkaProduct.GroupID,
	// 	conf.KafkaProduct.ClientID,
	// )

	// defer kafkaProductConsumer.Close()
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go kafkaProductConsumer.Consume(ctx, func(msg kafka.Message) error {
	// 	key := string(msg.Key)
	// 	return product.HandleProductEvent(msg.Value, key, productService)
	// })

	go func() {
		if err := router.Run(":8083"); err != nil {
			fmt.Println("Ошибка при запуске HTTP-сервера:", err)
		}
	}()

	select {}
}
