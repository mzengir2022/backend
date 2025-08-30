package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"my-project/config"
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		PhoneNumber: "09123456789",
		Email:       "test@example.com",
		Password:    "password",
	}
	jsonUser, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdUser models.User
	json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.Equal(t, user.PhoneNumber, createdUser.PhoneNumber)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Empty(t, createdUser.Password) // Password should not be in the response
}

func TestLogin(t *testing.T) {
	setupDatabase()

	// Create a user to test login
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)
	user := models.User{
		PhoneNumber: "09123456789",
		Email:       "test@example.com",
		Password:    hashedPassword,
		Role:        "user",
	}
	database.DB.Create(&user)

	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})

	r := setupRouter()
	r.POST("/login", Login)

	loginReq := LoginRequest{
		PhoneNumber: "09123456789",
		Password:    "password",
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

func TestAssignRole(t *testing.T) {
	setupDatabase()

	// Create an admin and a regular user
	admin := models.User{PhoneNumber: "09120000000", Email: "admin@example.com", Password: "password", Role: "admin"}
	user := models.User{PhoneNumber: "09121111111", Email: "user@example.com", Password: "password", Role: "user"}
	database.DB.Create(&admin)
	database.DB.Create(&user)

	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})
	adminToken, _ := auth.GenerateJWT(admin.PhoneNumber, admin.Role)
	userToken, _ := auth.GenerateJWT(user.PhoneNumber, user.Role)

	r := setupRouter()
	r.PUT("/users/:id/role", auth.AuthMiddleware(), auth.RoleAuthMiddleware("admin"), AssignRole)

	// Test with admin token (should succeed)
	newRole := "moderator"
	reqBody := AssignRoleRequest{Role: newRole}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/users/"+fmt.Sprintf("%d", user.ID)+"/role", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var updatedUser models.User
	database.DB.First(&updatedUser, user.ID)
	assert.Equal(t, newRole, updatedUser.Role)

	// Test with user token (should fail)
	req, _ = http.NewRequest("PUT", "/users/"+fmt.Sprintf("%d", admin.ID)+"/role", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequestEmailCode(t *testing.T) {
	setupDatabase()
	r := setupRouter()
	r.POST("/login/email/request", RequestEmailCode)

	user := models.User{
		PhoneNumber: "09123456789",
		Email:       "test@example.com",
		Password:    "password",
	}
	database.DB.Create(&user)

	reqBody := RequestEmailCodeRequest{Email: "test@example.com"}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login/email/request", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedUser models.User
	database.DB.First(&updatedUser, user.ID)
	assert.NotEmpty(t, updatedUser.EmailVerificationCode)
}

func TestVerifyEmailCode(t *testing.T) {
	setupDatabase()
	r := setupRouter()
	r.POST("/login/email/verify", VerifyEmailCode)

	code := "123456"
	user := models.User{
		PhoneNumber:                  "09123456789",
		Email:                        "test@example.com",
		Password:                     "password",
		EmailVerificationCode:        code,
		EmailVerificationCodeExpiresAt: time.Now().Add(5 * time.Minute),
	}
	database.DB.Create(&user)
	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})

	reqBody := VerifyEmailCodeRequest{Email: "test@example.com", Code: code}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login/email/verify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}

func TestRequestSMSCode(t *testing.T) {
	setupDatabase()
	r := setupRouter()
	r.POST("/login/sms/request", RequestSMSCode)

	user := models.User{
		PhoneNumber: "09123456789",
		Email:       "test@example.com",
		Password:    "password",
	}
	database.DB.Create(&user)

	reqBody := RequestCodeRequest{PhoneNumber: "09123456789"}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login/sms/request", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedUser models.User
	database.DB.First(&updatedUser, user.ID)
	assert.NotEmpty(t, updatedUser.VerificationCode)
}

func TestVerifySMSCode(t *testing.T) {
	setupDatabase()
	r := setupRouter()
	r.POST("/login/sms/verify", VerifySMSCode)

	code := "123456"
	user := models.User{
		PhoneNumber:               "09123456789",
		Email:                     "test@example.com",
		Password:                  "password",
		VerificationCode:          code,
		VerificationCodeExpiresAt: time.Now().Add(5 * time.Minute),
	}
	database.DB.Create(&user)
	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})

	reqBody := VerifyCodeRequest{PhoneNumber: "09123456789", Code: code}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login/sms/verify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}
