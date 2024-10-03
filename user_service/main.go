package main

import (
	"user_service/db"
	"user_service/handlers"
	"user_service/rabbitmq"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	db.Init()
	defer db.DB.Close()

	// Initialize RabbitMQ
	rabbitmq.Init()
	defer rabbitmq.Close()

	// Set up router
	r := gin.Default()

	r.POST("/register", handlers.RegisterUser)
	r.POST("/login", handlers.LoginUser)

	// Protected routes
	authorized := r.Group("/", handlers.Authenticate)
	{
		authorized.PUT("/profile", handlers.UpdateProfile)
		authorized.GET("/users", handlers.GetAllUsers)
		authorized.GET("/users/:id", handlers.GetUserByID)
	}

	// Start server
	r.Run(":8081")
}
