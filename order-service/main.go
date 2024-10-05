package main

import (
	"order-service/db"
	"order-service/handlers"
	"order-service/rabbitmq"
	"order-service/utils"

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

	// Initialize HTTP Client
	utils.InitHTTPClient()

	// Load existing products from Product Service
	rabbitmq.LoadExistingProducts()

	// Start listening for events
	rabbitmq.ListenForEvents()

	// Set up router
	r := gin.Default()

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register some custom metrics
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_http_requests_total",
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
