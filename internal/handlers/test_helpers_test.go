package handlers

import (
	"fmt"
	"my-project/config"
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes a test database and JWT
func setupTestDB() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	database.DB = db
	database.DB.AutoMigrate(&models.User{}, &models.Restaurant{}, &models.Menu{}, &models.MenuItem{})
	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})
}

// createTestUser creates a user for testing
func createTestUser(role string) (models.User, string) {
	uniqueID := time.Now().UnixNano()
	user := models.User{
		PhoneNumber: fmt.Sprintf("0912%d", uniqueID),
		Email:       fmt.Sprintf("%d@example.com", uniqueID),
		Password:    "password",
		Role:        role,
	}
	hashedPassword, _ := auth.HashPassword(user.Password)
	user.Password = hashedPassword
	database.DB.Create(&user)
	token, _ := auth.GenerateJWT(user.ID, user.PhoneNumber, user.Role)
	return user, token
}
