package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupUserTestRouter defines routes needed for user handler tests
func setupUserTestRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/signup", CreateUser)
	r.POST("/login", Login)
	r.POST("/login/sms/request", RequestSMSCode)
	r.POST("/login/sms/verify", VerifySMSCode)
	r.POST("/login/email/request", RequestEmailCode)
	r.POST("/login/email/verify", VerifyEmailCode)

	users := r.Group("/api/v1/users")
	users.Use(auth.AuthMiddleware())
	users.Use(auth.RoleAuthMiddleware("admin"))
	{
		users.PUT("/:id/role", AssignRole)
	}
	return r
}

func TestCreateUser(t *testing.T) {
	setupTestDB()
	r := setupUserTestRouter()

	reqBody := CreateUserRequest{
		PhoneNumber: "09123456789",
		Email:       "test@example.com",
		Password:    "password",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLogin(t *testing.T) {
	setupTestDB()
	r := setupUserTestRouter()

	password := "password"
	hashedPassword, _ := auth.HashPassword(password)
	user := models.User{
		PhoneNumber: "09123456789",
		Email:       "test@example.com",
		Password:    hashedPassword,
	}
	database.DB.Create(&user)

	reqBody := LoginRequest{
		PhoneNumber: "09123456789",
		Password:    "password",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}

func TestAssignRole(t *testing.T) {
	setupTestDB()
	r := setupUserTestRouter()

	_, adminToken := createTestUser("admin")
	user, _ := createTestUser("user")

	reqBody := AssignRoleRequest{Role: "moderator"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/users/"+fmt.Sprintf("%d", user.ID)+"/role", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var updatedUser models.User
	database.DB.First(&updatedUser, user.ID)
	assert.Equal(t, "moderator", updatedUser.Role)
}

// ... (other user tests can be added here, they were omitted for brevity in previous steps but the structure is the same)
