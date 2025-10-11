package migrations

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/favorites-service/internal/favorites"
)

func CheckForMigrations() error {

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		logger.Info("🚀 Starting migrations...")
		if err := RunMigrations(); err != nil {
			logger.Errorf("Error processing migrations: %v", err)
		}
		return nil
	}
	// if not "migrate" args[1]
	return nil
}

func RunMigrations() error {
    // Убираем godotenv, переменные уже должны быть в окружении контейнера
    dsn := os.Getenv("DSN")
    if dsn == "" {
        panic("DSN is not set in environment")
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        //DisableForeignKeyConstraintWhenMigrating: true,
    })
    if err != nil {
        panic(err)
    }

    err = db.AutoMigrate(favorites.Favorite{})
    if err != nil {
        return err
    }

    logger.Info("✅ Migrations completed")
    return nil
}
