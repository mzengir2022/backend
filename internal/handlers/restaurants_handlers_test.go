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

// setupTestRouter initializes a test router
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

// setupTestDB initializes a test database
func setupTestDB() {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	database.DB = db
	database.DB.AutoMigrate(&models.User{}, &models.Restaurant{}, &models.Menu{}, &models.MenuItem{})
	auth.InitializeJWT(&config.Config{JWTSecret: "test-secret"})
}

// createTestUser creates a user for testing
func createTestUser(role string) (models.User, string) {
	// Use current time to ensure unique phone number and email for each test run
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

func TestCreateRestaurant(t *testing.T) {
	setupTestDB()
	_, token := createTestUser("user")

	r := setupTestRouter()
	r.POST("/restaurants", auth.AuthMiddleware(), CreateRestaurant)

	restaurant := models.Restaurant{
		Name:    "Test Restaurant",
		Address: "123 Test St",
	}
	jsonRestaurant, _ := json.Marshal(restaurant)

	req, _ := http.NewRequest("POST", "/restaurants", bytes.NewBuffer(jsonRestaurant))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createdRestaurant models.Restaurant
	json.Unmarshal(w.Body.Bytes(), &createdRestaurant)
	assert.Equal(t, restaurant.Name, createdRestaurant.Name)
	assert.NotZero(t, createdRestaurant.UserID)
}

func TestUpdateRestaurant(t *testing.T) {
	setupTestDB()
	owner, ownerToken := createTestUser("user")
	_, otherToken := createTestUser("user")

	r := setupTestRouter()
	r.PUT("/restaurants/:id", auth.AuthMiddleware(), auth.OwnershipAuthMiddleware(), UpdateRestaurant)

	restaurant := models.Restaurant{
		Name:    "Original Name",
		Address: "Original Address",
		UserID:  owner.ID,
	}
	database.DB.Create(&restaurant)

	updatedInfo := models.Restaurant{Name: "New Name", Address: "New Address"}
	jsonBody, _ := json.Marshal(updatedInfo)

	// Test update by owner (should succeed)
	req, _ := http.NewRequest("PUT", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updatedRestaurant models.Restaurant
	json.Unmarshal(w.Body.Bytes(), &updatedRestaurant)
	assert.Equal(t, updatedInfo.Name, updatedRestaurant.Name)

	// Test update by non-owner (should fail)
	req, _ = http.NewRequest("PUT", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteRestaurant(t *testing.T) {
	setupTestDB()
	owner, ownerToken := createTestUser("user")
	_, otherToken := createTestUser("user")

	r := setupTestRouter()
	r.DELETE("/restaurants/:id", auth.AuthMiddleware(), auth.OwnershipAuthMiddleware(), DeleteRestaurant)

	restaurant := models.Restaurant{
		Name:    "To Be Deleted",
		Address: "Delete St",
		UserID:  owner.ID,
	}
	database.DB.Create(&restaurant)

	// Test delete by non-owner (should fail)
	req, _ := http.NewRequest("DELETE", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Test delete by owner (should succeed)
	req, _ = http.NewRequest("DELETE", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateMenu(t *testing.T) {
	setupTestDB()
	owner, ownerToken := createTestUser("user")
	_, otherToken := createTestUser("user")

	restaurant := models.Restaurant{UserID: owner.ID}
	database.DB.Create(&restaurant)

	r := setupTestRouter()
	r.POST("/restaurants/:id/menus", auth.AuthMiddleware(), auth.OwnershipAuthMiddleware(), CreateMenu)

	menu := models.Menu{Name: "Test Menu"}
	jsonBody, _ := json.Marshal(menu)

	// Test by owner (should succeed)
	req, _ := http.NewRequest("POST", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test by non-owner (should fail)
	req, _ = http.NewRequest("POST", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAddMenuItem(t *testing.T) {
	setupTestDB()
	owner, ownerToken := createTestUser("user")

	restaurant := models.Restaurant{UserID: owner.ID}
	database.DB.Create(&restaurant)
	menu := models.Menu{RestaurantID: restaurant.ID}
	database.DB.Create(&menu)

	r := setupTestRouter()
	r.POST("/menus/:menu_id/items", auth.AuthMiddleware(), auth.MenuOwnershipAuthMiddleware(), AddMenuItem)

	item := models.MenuItem{Name: "Test Item", Price: 9.99}
	jsonBody, _ := json.Marshal(item)

	req, _ := http.NewRequest("POST", "/menus/"+fmt.Sprintf("%d", menu.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSetDailyMenu(t *testing.T) {
	setupTestDB()
	owner, ownerToken := createTestUser("user")

	restaurant := models.Restaurant{UserID: owner.ID}
	database.DB.Create(&restaurant)
	menu := models.Menu{RestaurantID: restaurant.ID}
	database.DB.Create(&menu)

	r := setupTestRouter()
	r.PUT("/restaurants/:id/daily-menu", auth.AuthMiddleware(), auth.OwnershipAuthMiddleware(), SetDailyMenu)

	reqBody := SetDailyMenuRequest{MenuID: menu.ID}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updatedRestaurant models.Restaurant
	database.DB.First(&updatedRestaurant, restaurant.ID)
	assert.NotNil(t, updatedRestaurant.DailyMenuID)
	assert.Equal(t, menu.ID, *updatedRestaurant.DailyMenuID)
}
