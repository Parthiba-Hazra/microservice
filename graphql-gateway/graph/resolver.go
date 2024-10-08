package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"graphql-gateway/cache"
	"graphql-gateway/graph/model"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
)

type Resolver struct{}

// Query Resolver Implementation
func (r *Resolver) Users(ctx context.Context) ([]*model.User, error) {
	// Define the cache key
	cacheKey := "users"

	// Try to get data from Redis cache
	cachedData, err := cache.RedisClient.Get(cacheKey).Result()
	if err == nil {
		// Cache hit: unmarshal the cached data
		var users []*model.User
		if err := json.Unmarshal([]byte(cachedData), &users); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached users: %v", err)
		}
		fmt.Println("Cache hit: returning users from Redis")
		return users, nil
	} else if err != redis.Nil {
		// Redis error (other than key not found)
		return nil, fmt.Errorf("failed to get users from Redis: %v", err)
	}

	// Cache miss: fetch from User Service
	fmt.Println("Cache miss: fetching users from User Service")
	url := fmt.Sprintf("%s/users", getUserServiceURL())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(ctx))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve users: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string][]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	users, ok := result["users"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	var usersPtr []*model.User
	for _, user := range users {
		idFloat, ok := user["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type for id")
		}

		userModel := &model.User{
			ID:        fmt.Sprintf("%d", int(idFloat)),
			Username:  user["username"].(string),
			Email:     user["email"].(string),
			CreatedAt: user["created_at"].(string),
		}
		usersPtr = append(usersPtr, userModel)
	}

	// Store data in Redis cache for 5 minutes
	usersJSON, err := json.Marshal(usersPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal users for caching: %v", err)
	}
	err = cache.RedisClient.Set(cacheKey, usersJSON, 5*time.Minute).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to set users in Redis cache: %v", err)
	}

	return usersPtr, nil
}

func (r *Resolver) User(ctx context.Context, id string) (*model.User, error) {
	// Define the cache key for the specific user
	cacheKey := fmt.Sprintf("user:%s", id)

	// Try to get data from Redis cache
	cachedData, err := cache.RedisClient.Get(cacheKey).Result()
	if err == nil {
		// Cache hit: unmarshal the cached data
		var user model.User
		if err := json.Unmarshal([]byte(cachedData), &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached user: %v", err)
		}
		fmt.Println("Cache hit: returning user from Redis")
		return &user, nil
	} else if err != redis.Nil {
		// Redis error (other than key not found)
		return nil, fmt.Errorf("failed to get user from Redis: %v", err)
	}

	// Cache miss: fetch from User Service
	fmt.Println("Cache miss: fetching user from User Service")
	url := fmt.Sprintf("%s/users/%s", getUserServiceURL(), id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(ctx))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve user: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response into a map to extract the user data
	var result map[string]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	userData, ok := result["user"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	idFloat, ok := userData["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for id")
	}

	user := &model.User{
		ID:        fmt.Sprintf("%d", int(idFloat)),
		Username:  userData["username"].(string),
		Email:     userData["email"].(string),
		CreatedAt: userData["created_at"].(string),
	}

	// Store data in Redis cache for 5 minutes
	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user for caching: %v", err)
	}
	err = cache.RedisClient.Set(cacheKey, userJSON, 5*time.Minute).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to set user in Redis cache: %v", err)
	}

	return user, nil
}

func (r *Resolver) RegisterUser(ctx context.Context, input model.RegisterInput) (*model.RegisterUserResponse, error) {
	url := fmt.Sprintf("%s/register", getUserServiceURL())

	userData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(userData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to register user: %s", resp.Status)
	}

	var apiResponse map[string]string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	message, exists := apiResponse["message"]
	if !exists {
		return nil, fmt.Errorf("unexpected response format")
	}

	return &model.RegisterUserResponse{Message: message}, nil
}

// Helper function to get User Service URL
func getUserServiceURL() string {
	url := os.Getenv("USER_SERVICE_URL")
	if url == "" {
		url = "http://user-service:8081"
	}
	return url
}

func (r *Resolver) Products(ctx context.Context) ([]*model.Product, error) {
	// Define the cache key
	cacheKey := "products"

	// Try to get data from Redis cache
	cachedData, err := cache.RedisClient.Get(cacheKey).Result()
	if err == nil {
		// Cache hit: unmarshal the cached data
		var products []*model.Product
		if err := json.Unmarshal([]byte(cachedData), &products); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached products: %v", err)
		}
		fmt.Println("Cache hit: returning products from Redis")
		return products, nil
	} else if err != redis.Nil {
		// Redis error (other than key not found)
		return nil, fmt.Errorf("failed to get products from Redis: %v", err)
	}

	// Cache miss: fetch from Product Service
	fmt.Println("Cache miss: fetching products from Product Service")
	url := fmt.Sprintf("%s/products", getProductServiceURL())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve products: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response into a generic map to manipulate data types
	var result map[string][]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	productsData, ok := result["products"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	var products []*model.Product
	for _, product := range productsData {
		idFloat, ok := product["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type for product id")
		}
		productModel := &model.Product{
			ID:          fmt.Sprintf("%d", int(idFloat)),
			Name:        product["name"].(string),
			Description: product["description"].(string),
			Price:       product["price"].(float64),
			Inventory:   int(product["inventory"].(float64)),
			CreatedAt:   product["created_at"].(string),
		}
		products = append(products, productModel)
	}

	// Store data in Redis cache for 5 minutes
	productsJSON, err := json.Marshal(products)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal products for caching: %v", err)
	}
	err = cache.RedisClient.Set(cacheKey, productsJSON, 5*time.Minute).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to set products in Redis cache: %v", err)
	}

	return products, nil
}

func (r *Resolver) Product(ctx context.Context, id string) (*model.Product, error) {
	// Define the cache key for the specific product
	cacheKey := fmt.Sprintf("product:%s", id)

	// Try to get data from Redis cache
	cachedData, err := cache.RedisClient.Get(cacheKey).Result()
	if err == nil {
		// Cache hit: unmarshal the cached data
		var product *model.Product
		if err := json.Unmarshal([]byte(cachedData), &product); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached product: %v", err)
		}
		fmt.Println("Cache hit: returning product from Redis")
		return product, nil
	} else if err != redis.Nil {
		// Redis error (other than key not found)
		return nil, fmt.Errorf("failed to get product from Redis: %v", err)
	}

	// Cache miss: fetch from Product Service
	fmt.Println("Cache miss: fetching product from Product Service")
	url := fmt.Sprintf("%s/products/%s", getProductServiceURL(), id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve product: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response into a generic map to manipulate data types
	var result map[string]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	productData, ok := result["product"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	idFloat, ok := productData["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for product id")
	}

	product := &model.Product{
		ID:          fmt.Sprintf("%d", int(idFloat)),
		Name:        productData["name"].(string),
		Description: productData["description"].(string),
		Price:       productData["price"].(float64),
		Inventory:   int(productData["inventory"].(float64)),
		CreatedAt:   productData["created_at"].(string),
	}

	// Store data in Redis cache for 5 minutes
	productJSON, err := json.Marshal(product)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal product for caching: %v", err)
	}
	err = cache.RedisClient.Set(cacheKey, productJSON, 5*time.Minute).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to set product in Redis cache: %v", err)
	}

	return product, nil
}

func (r *Resolver) CreateProduct(ctx context.Context, input model.ProductInput) (*model.ProductResponse, error) {
	url := fmt.Sprintf("%s/products", getProductServiceURL())

	productData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(productData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Get JWT token from context
	authHeader := getAuthHeader(ctx)
	if authHeader == "" {
		return nil, fmt.Errorf("unauthorized: missing Authorization header")
	}

	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create product: %s", resp.Status)
	}

	var apiResponse map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	// Convert "message" field to string
	messageInterface, exists := apiResponse["message"]
	if !exists {
		return nil, fmt.Errorf("unexpected response format: 'message' key not found")
	}

	message, ok := messageInterface.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected response format: 'message' value is not a string")
	}

	return &model.ProductResponse{Message: message}, nil
}

// Helper function to get Product Service URL
func getProductServiceURL() string {
	url := os.Getenv("PRODUCT_SERVICE_URL")
	if url == "" {
		url = "http://product-service:8082"
	}
	return url
}

// Helper function to get authorization header from context
func getAuthHeader(ctx context.Context) string {
	authHeader, present := ctx.Value("Authorization").(string)
	if !present {
		log.Printf("no authorization %v, %v\n", present, authHeader)
	}
	return authHeader
}

func (r *Resolver) Orders(ctx context.Context) ([]*model.Order, error) {
	url := fmt.Sprintf("%s/orders", getOrderServiceURL())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(ctx))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve orders: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string][]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	ordersData, ok := result["orders"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	var orders []*model.Order
	for _, order := range ordersData {
		idFloat, ok := order["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type for order id")
		}

		orderModel := &model.Order{
			ID:        fmt.Sprintf("%d", int(idFloat)),
			UserID:    int(order["user_id"].(float64)),
			Status:    order["status"].(string),
			Total:     order["total"].(float64),
			CreatedAt: order["created_at"].(string),
		}

		// Parsing items
		items := order["items"].([]interface{})
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			itemID, _ := itemMap["id"].(float64)
			orderModel.Items = append(orderModel.Items, &model.OrderItem{
				ID:        fmt.Sprintf("%d", int(itemID)),
				OrderID:   int(itemMap["order_id"].(float64)),
				ProductID: int(itemMap["product_id"].(float64)),
				Quantity:  int(itemMap["quantity"].(float64)),
				Price:     itemMap["price"].(float64),
			})
		}

		orders = append(orders, orderModel)
	}

	return orders, nil
}

func (r *Resolver) Order(ctx context.Context, id string) (*model.Order, error) {
	url := fmt.Sprintf("%s/orders/%s", getOrderServiceURL(), id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(ctx))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve order: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	orderData, ok := result["order"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	idFloat, ok := orderData["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for order id")
	}

	order := &model.Order{
		ID:        fmt.Sprintf("%d", int(idFloat)),
		UserID:    int(orderData["user_id"].(float64)),
		Status:    orderData["status"].(string),
		Total:     orderData["total"].(float64),
		CreatedAt: orderData["created_at"].(string),
	}

	// Parsing items
	items := orderData["items"].([]interface{})
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		itemID, _ := itemMap["id"].(float64)
		order.Items = append(order.Items, &model.OrderItem{
			ID:        fmt.Sprintf("%d", int(itemID)),
			OrderID:   int(itemMap["order_id"].(float64)),
			ProductID: int(itemMap["product_id"].(float64)),
			Quantity:  int(itemMap["quantity"].(float64)),
			Price:     itemMap["price"].(float64),
		})
	}

	return order, nil
}

func (r *Resolver) PlaceOrder(ctx context.Context, input model.OrderInput) (*model.OrderResponse, error) {
	url := fmt.Sprintf("%s/orders", getOrderServiceURL())

	orderData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(orderData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Get JWT token from context
	authHeader := getAuthHeader(ctx)
	if authHeader == "" {
		return nil, fmt.Errorf("unauthorized: missing Authorization header")
	}

	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResponse map[string]interface{}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			return nil, err
		}

		message, _ := apiResponse["error"].(string)
		return nil, fmt.Errorf("failed to place order: %s", message)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	message, exists := apiResponse["message"].(string)
	if !exists {
		return nil, fmt.Errorf("unexpected response format: 'message' key not found")
	}

	orderID, _ := apiResponse["order_id"].(float64)

	return &model.OrderResponse{
		Message: message,
		OrderID: fmt.Sprintf("%d", int(orderID)),
	}, nil
}

func getOrderServiceURL() string {
	url := os.Getenv("ORDER_SERVICE_URL")
	if url == "" {
		url = "http://order-service:8083"
	}
	return url
}
