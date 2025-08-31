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

// setupRestaurantTestRouter defines routes needed for restaurant handler tests
func setupRestaurantTestRouter() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v1")
	api.Use(auth.AuthMiddleware())
	{
		restaurants := api.Group("/restaurants")
		{
			restaurants.POST("", CreateRestaurant)
			restaurants.PUT("/:id", auth.OwnershipAuthMiddleware(), UpdateRestaurant)
			restaurants.DELETE("/:id", auth.OwnershipAuthMiddleware(), DeleteRestaurant)
			restaurants.POST("/:id/menus", auth.OwnershipAuthMiddleware(), CreateMenu)
			restaurants.PUT("/:id/daily-menu", auth.OwnershipAuthMiddleware(), SetDailyMenu)
		}
		menus := api.Group("/menus")
		{
			menus.POST("/:menu_id/items", auth.MenuOwnershipAuthMiddleware(), AddMenuItem)
		}
	}
	return r
}


func TestCreateRestaurant(t *testing.T) {
	setupTestDB()
	r := setupRestaurantTestRouter()
	user, token := createTestUser("user")

	reqBody := CreateRestaurantRequest{
		Name:    "New Test Restaurant",
		Address: "456 New St",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/restaurants", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createdRestaurant models.Restaurant
	json.Unmarshal(w.Body.Bytes(), &createdRestaurant)
	assert.Equal(t, reqBody.Name, createdRestaurant.Name)
	assert.Equal(t, user.ID, createdRestaurant.UserID)
}

func TestUpdateRestaurant(t *testing.T) {
	setupTestDB()
	r := setupRestaurantTestRouter()
	owner, ownerToken := createTestUser("user")
	_, otherToken := createTestUser("user")

	restaurant := models.Restaurant{UserID: owner.ID, Name: "Old Name"}
	database.DB.Create(&restaurant)

	reqBody := UpdateRestaurantRequest{Name: "Updated Name"}
	jsonBody, _ := json.Marshal(reqBody)

	// Test update by owner (should succeed)
	req, _ := http.NewRequest("PUT", "/api/v1/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test update by non-owner (should fail)
	req, _ = http.NewRequest("PUT", "/api/v1/restaurants/"+fmt.Sprintf("%d", restaurant.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateMenu(t *testing.T) {
	setupTestDB()
	r := setupRestaurantTestRouter()
	owner, ownerToken := createTestUser("user")
	restaurant := models.Restaurant{UserID: owner.ID}
	database.DB.Create(&restaurant)

	reqBody := CreateMenuRequest{Name: "New Menu"}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/restaurants/"+fmt.Sprintf("%d", restaurant.ID)+"/menus", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAddMenuItem(t *testing.T) {
	setupTestDB()
	r := setupRestaurantTestRouter()
	owner, ownerToken := createTestUser("user")
	restaurant := models.Restaurant{UserID: owner.ID}
	database.DB.Create(&restaurant)
	menu := models.Menu{RestaurantID: restaurant.ID, Name: "Test Menu"}
	database.DB.Create(&menu)

	reqBody := AddMenuItemRequest{Name: "Test Item", Price: 9.99}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/menus/"+fmt.Sprintf("%d", menu.ID)+"/items", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSetDailyMenu(t *testing.T) {
	setupTestDB()
	r := setupRestaurantTestRouter()
	owner, ownerToken := createTestUser("user")
	restaurant := models.Restaurant{UserID: owner.ID}
	database.DB.Create(&restaurant)
	menu := models.Menu{RestaurantID: restaurant.ID, Name: "Daily Special"}
	database.DB.Create(&menu)

	reqBody := SetDailyMenuRequest{MenuID: menu.ID}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/v1/restaurants/"+fmt.Sprintf("%d", restaurant.ID)+"/daily-menu", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updatedRestaurant models.Restaurant
	database.DB.First(&updatedRestaurant, restaurant.ID)
	assert.NotNil(t, updatedRestaurant.DailyMenuID)
	assert.Equal(t, menu.ID, *updatedRestaurant.DailyMenuID)
}
