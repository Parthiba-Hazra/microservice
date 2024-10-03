package main

import (
	"product-service/db"
	"product-service/handlers"
	"product-service/rabbitmq"

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
	rabbitmq.ListenForOrderPlacedEvents()

	// Set up router
	r := gin.Default()

	// Public routes
	r.GET("/products", handlers.GetAllProducts)
	r.GET("/products/:id", handlers.GetProductByID)

	// Protected routes
	authorized := r.Group("/", handlers.Authenticate)
	{
		authorized.POST("/products", handlers.CreateProduct)
		authorized.PUT("/products/:id", handlers.UpdateProduct)
		authorized.DELETE("/products/:id", handlers.DeleteProduct)
	}

	// Start server
	r.Run(":8082")
}
