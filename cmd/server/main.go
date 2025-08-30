package main

import (
	"my-project/config"
	_ "my-project/docs" // This line is important for swag
	"my-project/internal/auth"
	"my-project/internal/database"
	"my-project/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           My Project API
// @version         1.0
// @description     This is a sample server for a Go project.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apiKey  ApiKeyAuth
// @in                          header
// @name                        Authorization
func main() {
	cfg := config.LoadConfig()
	database.Connect(cfg)
	auth.InitializeJWT(cfg)

	r := gin.Default()

	r.POST("/signup", handlers.CreateUser)
	r.POST("/login", handlers.Login)
	r.POST("/login/sms/request", handlers.RequestSMSCode)
	r.POST("/login/sms/verify", handlers.VerifySMSCode)
	r.POST("/login/email/request", handlers.RequestEmailCode)
	r.POST("/login/email/verify", handlers.VerifyEmailCode)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
