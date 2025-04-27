package configs

import (
	"os"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	Db DbConfig
	// Kafka KafkaConfig
}

type DbConfig struct {
	Dsn string
}

// type KafkaConfig struct {
// 	Brokers []string
// 	Topic   string
// 	GroupID string
// 	ClientID string
// }

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file, using default config", err.Error())
	}

	// brokersRaw := os.Getenv("KAFKA_BROKERS")
	// brokers := strings.Split(brokersRaw, ",")

	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		// Kafka: KafkaConfig{
		// 	Brokers: brokers,
		// 	Topic:   os.Getenv("KAFKA_TOPIC"),
		// 	GroupID: os.Getenv("KAFKA_GROUP_ID"),
		// 	ClientID: os.Getenv("KAFKA_CLIENT_ID"),
		// },
	}
}
