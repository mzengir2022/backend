package main

import (
	"my-project/config"
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	database.Connect(cfg)
	auth.InitializeJWT(cfg)

	r := gin.Default()

	r.POST("/signup", handlers.CreateUser)
	r.POST("/login", handlers.Login)

	api := r.Group("/api/v1")
	api.Use(auth.AuthMiddleware())
	{
		users := api.Group("/users")
		users.Use(auth.RoleAuthMiddleware("admin"))
		{
			users.GET("", handlers.GetUsers)
			users.GET("/:id", handlers.GetUser)
			users.PUT("/:id", handlers.UpdateUser)
			users.DELETE("/:id", handlers.DeleteUser)
		}
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
