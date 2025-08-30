package handlers

import (
	"bytes"
	"encoding/json"
	"my-project/config"
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

func setupDatabase() {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	database.DB = db
	database.DB.AutoMigrate(&models.User{})
}

func TestCreateUser(t *testing.T) {
	setupDatabase()
	r := setupRouter()
	r.POST("/signup", CreateUser)

	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
	}
	jsonUser, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdUser models.User
	json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.Equal(t, user.Username, createdUser.Username)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Empty(t, createdUser.Password) // Password should not be in the response
}

func TestLogin(t *testing.T) {
	setupDatabase()

	// Create a user to test login
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "user",
	}
	database.DB.Create(&user)

	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})

	r := setupRouter()
	r.POST("/login", Login)

	loginReq := LoginRequest{
		Username: "testuser",
		Password: "password",
	}
	jsonLogin, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonLogin))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}

func TestGetUsers(t *testing.T) {
	setupDatabase()

	// Create an admin and a regular user
	admin := models.User{Username: "admin", Email: "admin@example.com", Password: "password", Role: "admin"}
	user := models.User{Username: "user", Email: "user@example.com", Password: "password", Role: "user"}
	database.DB.Create(&admin)
	database.DB.Create(&user)

	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})
	adminToken, _ := auth.GenerateJWT(admin.Username, admin.Role)
	userToken, _ := auth.GenerateJWT(user.Username, user.Role)

	r := setupRouter()
	r.GET("/users", auth.AuthMiddleware(), auth.RoleAuthMiddleware("admin"), GetUsers)

	// Test with admin token (should succeed)
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []models.User
	json.Unmarshal(w.Body.Bytes(), &users)
	assert.Len(t, users, 2)

	// Test with user token (should fail)
	req, _ = http.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
