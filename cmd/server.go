package main

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/db"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/favorites-service/internal/favorites"
	"github.com/ShopOnGO/favorites-service/migrations"
	"github.com/gorilla/mux"
)

func main() {
	migrations.CheckForMigrations()
	conf := configs.LoadConfig()
	database := db.NewDB(conf)
	router := mux.NewRouter()

	// repository
	favoriteRepo := favorites.NewFavoriteRepository(database)

	// service
	favoriteService := favorites.NewFavoriteService(favoriteRepo)

	// handler
	favorites.NewFavoriteHandler(router, favorites.FavoriteHandlerDeps{
		Config: conf,
		FavoriteService: favoriteService,
	})

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
		logger.Info("Favorites service listening on 8083")
		if err := http.ListenAndServe(":8083", router); err != nil {
			logger.Errorf("Failed to start HTTP server: %v", err)
		}
	}()

	select {}
}
