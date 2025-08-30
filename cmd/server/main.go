package main

import (
	"my-project/config"
	_ "my-project/docs" // This line is important for swag
	"my-project/internal/auth"
	"my-project/internal/database"
)

// @title           My Project API
// @version         1.0
// @description     This is a sample server for a Go project with restaurant menu features.
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

	r := setupRouter()
	r.Run() // listen and serve on 0.0.0.0:8080
}
