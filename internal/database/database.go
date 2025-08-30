package database

import (
	"log"
	"my-project/config"
	"my-project/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg *config.Config) {
	var err error
	dsn := cfg.DSN()
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connection established")

	// Auto-migrate the schema
	DB.AutoMigrate(&models.User{}, &models.Restaurant{}, &models.Menu{}, &models.MenuItem{})
	log.Println("Database schema migrated")
}
