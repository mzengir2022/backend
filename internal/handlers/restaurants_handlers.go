package handlers

import (
	"my-project/internal/database"
	"my-project/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// DTOs for Restaurant Handlers
type CreateRestaurantRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
}
type UpdateRestaurantRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// CreateRestaurant godoc
// @Summary      Create a new restaurant
// @Description  Creates a new restaurant for the authenticated user
// @Tags         restaurants
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        restaurant  body      CreateRestaurantRequest  true  "Restaurant info"
// @Success      201         {object}  models.Restaurant
// @Failure      400         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /api/v1/restaurants [post]
func CreateRestaurant(c *gin.Context) {
	var req CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	restaurant := models.Restaurant{
		Name:    req.Name,
		Address: req.Address,
		UserID:  userID.(uint),
	}

	if err := database.DB.Create(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create restaurant"})
		return
	}

	c.JSON(http.StatusCreated, restaurant)
}

// GetRestaurant godoc
// @Summary      Get a restaurant by ID
// @Description  Get details for a single restaurant
// @Tags         restaurants
// @Produce      json
// @Param        id   path      int  true  "Restaurant ID"
// @Success      200  {object}  models.Restaurant
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/restaurants/{id} [get]
func GetRestaurant(c *gin.Context) {
	id := c.Param("id")
	var restaurant models.Restaurant
	// Preload menus and menu items
	if err := database.DB.Preload("Menus.Items").First(&restaurant, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}
	c.JSON(http.StatusOK, restaurant)
}

// UpdateRestaurant godoc
// @Summary      Update a restaurant
// @Description  Update a restaurant's information (owner or admin only)
// @Tags         restaurants
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id          path      int                      true  "Restaurant ID"
// @Param        restaurant  body      UpdateRestaurantRequest  true  "Restaurant info"
// @Success      200         {object}  models.Restaurant
// @Failure      400         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /api/v1/restaurants/{id} [put]
func UpdateRestaurant(c *gin.Context) {
	id := c.Param("id")
	var restaurant models.Restaurant
	if err := database.DB.First(&restaurant, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	var req UpdateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		restaurant.Name = req.Name
	}
	if req.Address != "" {
		restaurant.Address = req.Address
	}

	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update restaurant"})
		return
	}
	c.JSON(http.StatusOK, restaurant)
}

// DeleteRestaurant godoc
// @Summary      Delete a restaurant
// @Description  Delete a restaurant by its ID (owner or admin only)
// @Tags         restaurants
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Restaurant ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/restaurants/{id} [delete]
func DeleteRestaurant(c *gin.Context) {
	id := c.Param("id")
	var restaurant models.Restaurant
	if err := database.DB.First(&restaurant, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	if err := database.DB.Delete(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete restaurant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Restaurant deleted successfully"})
}

// Menu Handlers
type CreateMenuRequest struct {
	Name string `json:"name" binding:"required"`
}
type UpdateMenuRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateMenu godoc
// @Summary      Create a new menu
// @Description  Creates a new menu for a restaurant (owner or admin only)
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id    path      int                true  "Restaurant ID"
// @Param        menu  body      CreateMenuRequest  true  "Menu info"
// @Success      201   {object}  models.Menu
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/restaurants/{id}/menus [post]
func CreateMenu(c *gin.Context) {
	restaurantID := Atoi(c.Param("id"))
	var req CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	menu := models.Menu{
		Name:         req.Name,
		RestaurantID: restaurantID,
	}

	if err := database.DB.Create(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
		return
	}

	c.JSON(http.StatusCreated, menu)
}

// MenuItem Handlers
type AddMenuItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required"`
}
type UpdateMenuItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// AddMenuItem godoc
// @Summary      Add a menu item
// @Description  Adds a new item to a menu (owner or admin only)
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        menu_id  path      int                 true  "Menu ID"
// @Param        item     body      AddMenuItemRequest  true  "Menu item info"
// @Success      201      {object}  models.MenuItem
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/menus/{menu_id}/items [post]
func AddMenuItem(c *gin.Context) {
	menuID := Atoi(c.Param("menu_id"))
	var req AddMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.MenuItem{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		MenuID:      menuID,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add menu item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GetMenu godoc
// @Summary      Get a menu by ID
// @Description  Get details for a single menu, including its items
// @Tags         menus
// @Produce      json
// @Param        menu_id  path      int  true  "Menu ID"
// @Success      200      {object}  models.Menu
// @Failure      404      {object}  map[string]string
// @Router       /api/v1/menus/{menu_id} [get]
func GetMenu(c *gin.Context) {
	menuID := c.Param("menu_id")
	var menu models.Menu
	if err := database.DB.Preload("Items").First(&menu, menuID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}
	c.JSON(http.StatusOK, menu)
}

// UpdateMenu godoc
// @Summary      Update a menu
// @Description  Update a menu's name (owner or admin only)
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        menu_id  path      int                true  "Menu ID"
// @Param        menu     body      UpdateMenuRequest  true  "Menu info"
// @Success      200      {object}  models.Menu
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/menus/{menu_id} [put]
func UpdateMenu(c *gin.Context) {
	menuID := c.Param("menu_id")
	var menu models.Menu
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	var req UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	menu.Name = req.Name

	if err := database.DB.Save(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
		return
	}
	c.JSON(http.StatusOK, menu)
}

// DeleteMenu godoc
// @Summary      Delete a menu
// @Description  Delete a menu by its ID (owner or admin only)
// @Tags         menus
// @Produce      json
// @Security     ApiKeyAuth
// @Param        menu_id  path      int  true  "Menu ID"
// @Success      200      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/menus/{menu_id} [delete]
func DeleteMenu(c *gin.Context) {
	menuID := c.Param("menu_id")
	var menu models.Menu
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	if err := database.DB.Delete(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Menu deleted successfully"})
}

// UpdateMenuItem godoc
// @Summary      Update a menu item
// @Description  Update a menu item's information (owner or admin only)
// @Tags         menu-items
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        item_id  path      int                    true  "Menu Item ID"
// @Param        item     body      UpdateMenuItemRequest  true  "Menu item info"
// @Success      200      {object}  models.MenuItem
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/menu-items/{item_id} [put]
func UpdateMenuItem(c *gin.Context) {
	itemID := c.Param("item_id")
	var item models.MenuItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	var req UpdateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Price != 0 {
		item.Price = req.Price
	}

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteMenuItem godoc
// @Summary      Delete a menu item
// @Description  Delete a menu item by its ID (owner or admin only)
// @Tags         menu-items
// @Produce      json
// @Security     ApiKeyAuth
// @Param        item_id  path      int  true  "Menu Item ID"
// @Success      200      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/menu-items/{item_id} [delete]
func DeleteMenuItem(c *gin.Context) {
	itemID := c.Param("item_id")
	var item models.MenuItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Menu item deleted successfully"})
}

// QRCode Handlers

// GenerateQRCode godoc
// @Summary      Generate a QR code for a restaurant's menu
// @Description  Generates and returns a QR code image linking to the restaurant's menu URL
// @Tags         restaurants
// @Produce      image/png
// @Param        id   path      int  true  "Restaurant ID"
// @Success      200  {string}  string "QR code image"
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/restaurants/{id}/qrcode [get]
func GenerateQRCode(c *gin.Context) {
	restaurantID := c.Param("id")
	// This URL should point to your frontend application's menu page
	url := "http://localhost:3000/restaurants/" + restaurantID + "/menu"

	qrCode, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", qrCode)
}

// DailyMenu Handlers

type SetDailyMenuRequest struct {
	MenuID uint `json:"menu_id" binding:"required"`
}

// SetDailyMenu godoc
// @Summary      Set the daily menu for a restaurant
// @Description  Sets a specific menu as the daily menu for a restaurant (owner or admin only)
// @Tags         restaurants
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id          path      int                  true  "Restaurant ID"
// @Param        daily_menu  body      SetDailyMenuRequest  true  "Daily Menu info"
// @Success      200         {object}  map[string]string
// @Failure      400         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /api/v1/restaurants/{id}/daily-menu [put]
func SetDailyMenu(c *gin.Context) {
	restaurantID := Atoi(c.Param("id"))
	var req SetDailyMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify menu belongs to the restaurant
	var menu models.Menu
	if err := database.DB.First(&menu, req.MenuID).Error; err != nil || menu.RestaurantID != restaurantID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Menu does not belong to this restaurant"})
		return
	}

	var restaurant models.Restaurant
	if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	restaurant.DailyMenuID = &req.MenuID
	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set daily menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Daily menu set successfully"})
}


// Atoi is a helper function to convert string to uint
func Atoi(s string) uint {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint(i)
}
