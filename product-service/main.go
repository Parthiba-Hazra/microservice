package main

import (
	"product-service/db"
	"product-service/handlers"
	"product-service/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register some custom metrics
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	prometheus.MustRegister(requestCounter)

	// Middleware to track requests
	r.Use(func(c *gin.Context) {
		requestCounter.With(prometheus.Labels{"method": c.Request.Method, "endpoint": c.FullPath()}).Inc()
		c.Next()
	})

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
