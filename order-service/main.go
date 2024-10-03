package main

import (
	"order-service/db"
	"order-service/handlers"
	"order-service/rabbitmq"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	db.Init()
	defer db.DB.Close()

	// Initialize RabbitMQ
	rabbitmq.Init()
	defer rabbitmq.Close()

	// Start listening for events
	rabbitmq.ListenForEvents()

	// Set up router
	r := gin.Default()

	// Protected routes
	authorized := r.Group("/", handlers.Authenticate)
	{
		authorized.POST("/orders", handlers.PlaceOrder)
		authorized.GET("/orders", handlers.GetAllOrders)
		authorized.GET("/orders/:id", handlers.GetOrderByID)
	}

	// Start server
	r.Run(":8083")
}
