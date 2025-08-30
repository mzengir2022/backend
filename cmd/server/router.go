package main

import (
	"my-project/internal/auth"
	"my-project/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Public routes
	r.POST("/signup", handlers.CreateUser)
	r.POST("/login", handlers.Login)
	r.POST("/login/sms/request", handlers.RequestSMSCode)
	r.POST("/login/sms/verify", handlers.VerifySMSCode)
	r.POST("/login/email/request", handlers.RequestEmailCode)
	r.POST("/login/email/verify", handlers.VerifyEmailCode)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// API v1 group
	api := r.Group("/api/v1")
	api.Use(auth.AuthMiddleware())
	{
		// User management routes (admin only)
		users := api.Group("/users")
		users.Use(auth.RoleAuthMiddleware("admin"))
		{
			users.GET("", handlers.GetUsers)
			users.GET("/:id", handlers.GetUser)
			users.PUT("/:id", handlers.UpdateUser)
			users.DELETE("/:id", handlers.DeleteUser)
			users.PUT("/:id/role", handlers.AssignRole)
		}

		// Restaurant routes
		restaurants := api.Group("/restaurants")
		{
			restaurants.POST("", handlers.CreateRestaurant)
			restaurants.GET("/:id", handlers.GetRestaurant) // Publicly viewable restaurant details
			restaurants.PUT("/:id", auth.OwnershipAuthMiddleware(), handlers.UpdateRestaurant)
			restaurants.DELETE("/:id", auth.OwnershipAuthMiddleware(), handlers.DeleteRestaurant)
			restaurants.GET("/:id/qrcode", handlers.GenerateQRCode) // Publicly get QR code
			restaurants.PUT("/:id/daily-menu", auth.OwnershipAuthMiddleware(), handlers.SetDailyMenu)

			// Nested Menu routes
			restaurants.POST("/:id/menus", auth.OwnershipAuthMiddleware(), handlers.CreateMenu)
		}

		// Menu routes
		menus := api.Group("/menus")
		{
			menus.GET("/:menu_id", handlers.GetMenu) // Publicly viewable menu details
			menus.PUT("/:menu_id", auth.MenuOwnershipAuthMiddleware(), handlers.UpdateMenu)
			menus.DELETE("/:menu_id", auth.MenuOwnershipAuthMiddleware(), handlers.DeleteMenu)

			// Nested MenuItem routes
			menus.POST("/:menu_id/items", auth.MenuOwnershipAuthMiddleware(), handlers.AddMenuItem)
		}

		// MenuItem routes
		menuItems := api.Group("/menu-items")
		{
			menuItems.PUT("/:item_id", auth.MenuItemOwnershipAuthMiddleware(), handlers.UpdateMenuItem)
			menuItems.DELETE("/:item_id", auth.MenuItemOwnershipAuthMiddleware(), handlers.DeleteMenuItem)
		}
	}

	return r
}
