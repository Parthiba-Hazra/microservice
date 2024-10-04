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
	r.GET("/users", handlers.GetAllUsers)
	r.GET("/users/:id", handlers.GetUserByID)

	// Protected routes
	authorized := r.Group("/", handlers.Authenticate)
	{
		authorized.PUT("/profile", handlers.UpdateProfile)
	}

	// Start server
	r.Run(":8081")
}
