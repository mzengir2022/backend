package auth

import (
	"my-project/internal/database"
	"my-project/internal/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("phone_number", claims.PhoneNumber)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func MenuItemOwnershipAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User ID not found in context"})
			return
		}

		userRole, _ := c.Get("role")
		if userRole == "admin" {
			c.Next()
			return
		}

		itemID := c.Param("item_id")
		var ownerID uint
		err := database.DB.Table("menu_items").
			Select("restaurants.user_id").
			Joins("join menus on menus.id = menu_items.menu_id").
			Joins("join restaurants on restaurants.id = menus.restaurant_id").
			Where("menu_items.id = ?", itemID).
			Row().
			Scan(&ownerID)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Menu item not found or ownership could not be verified"})
			return
		}

		if ownerID != userID.(uint) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this menu item"})
			return
		}

		c.Next()
	}
}

func MenuOwnershipAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User ID not found in context"})
			return
		}

		userRole, _ := c.Get("role")
		if userRole == "admin" {
			c.Next()
			return
		}

		menuID := c.Param("menu_id")
		var ownerID uint
		err := database.DB.Table("menus").
			Select("restaurants.user_id").
			Joins("join restaurants on restaurants.id = menus.restaurant_id").
			Where("menus.id = ?", menuID).
			Row().
			Scan(&ownerID)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Menu not found or ownership could not be verified"})
			return
		}

		if ownerID != userID.(uint) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this menu"})
			return
		}

		c.Next()
	}
}

func OwnershipAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User ID not found in context"})
			return
		}

		userRole, _ := c.Get("role")
		if userRole == "admin" {
			c.Next()
			return
		}

		restaurantID := c.Param("id")
		var restaurant models.Restaurant
		if err := database.DB.First(&restaurant, restaurantID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
			return
		}

		if restaurant.UserID != userID.(uint) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this restaurant"})
			return
		}

		c.Next()
	}
}

func RoleAuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User role not found in context"})
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role format in context"})
			return
		}

		if userRole != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not authorized to perform this action"})
			return
		}

		c.Next()
	}
}
