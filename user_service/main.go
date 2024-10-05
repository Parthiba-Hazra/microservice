package main

import (
	"user_service/db"
	"user_service/handlers"
	"user_service/rabbitmq"

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

	// Set up router
	r := gin.Default()

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register some custom metrics
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_http_requests_total",
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
