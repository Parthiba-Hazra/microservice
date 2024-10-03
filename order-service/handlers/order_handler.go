package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"order-service/db"
	"order-service/models"
	"order-service/rabbitmq"
	"order-service/utils"

	"strconv"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func Authenticate(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		c.Abort()
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
		c.Abort()
		return
	}
	tokenString := parts[1]

	claims := &utils.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return utils.JwtKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	// Store user information in context
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Next()
}

var mutex = &sync.RWMutex{}

func PlaceOrder(c *gin.Context) {
	// Authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var input models.OrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate total and prepare order items
	total := 0.0
	orderItems := []models.OrderItem{}
	for _, item := range input.Items {
		// Get product details
		product, ok := getProductDetails(item.ProductID)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Product %d not found", item.ProductID)})
			return
		}
		itemTotal := product.Price * float64(item.Quantity)
		total += itemTotal
		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Insert order
	var orderID int
	query := `INSERT INTO orders (user_id, status, total) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(query, userID, "Placed", total).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Insert order items
	for _, item := range orderItems {
		query = `INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(query, orderID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order items"})
			return
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Emit "Order Placed" event
	orderEvent := models.OrderPlacedEvent{
		OrderID: orderID,
		UserID:  userID.(int),
		Items:   []models.OrderItemInfo{},
	}
	for _, item := range orderItems {
		orderEvent.Items = append(orderEvent.Items, models.OrderItemInfo{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	rabbitmq.EmitOrderPlaced(orderEvent)

	c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully", "order_id": orderID})
}

func GetAllOrders(c *gin.Context) {
	// Authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := `SELECT id, user_id, status, total, created_at FROM orders WHERE user_id = $1`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Total, &order.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order"})
			return
		}

		// Get order items
		items, err := getOrderItems(order.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order items"})
			return
		}
		order.Items = items

		orders = append(orders, order)
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func GetOrderByID(c *gin.Context) {
	// Authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idParam := c.Param("id")
	orderID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	query := `SELECT id, user_id, status, total, created_at FROM orders WHERE id = $1`
	var order models.Order
	err = db.DB.QueryRow(query, orderID).Scan(&order.ID, &order.UserID, &order.Status, &order.Total, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		}
		return
	}

	// Check if the order belongs to the authenticated user
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get order items
	items, err := getOrderItems(order.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order items"})
		return
	}
	order.Items = items

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func getOrderItems(orderID int) ([]models.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := db.DB.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.OrderItem{}
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func getProductDetails(productID int) (models.Product, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	product, exists := rabbitmq.ProductCatalog[productID]
	return product, exists
}
