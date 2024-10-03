package rabbitmq

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"order-service/models"
	"order-service/utils"
	"sync"

	"github.com/streadway/amqp"
)

func EmitOrderPlaced(order models.OrderPlacedEvent) {
	body, err := json.Marshal(order)
	if err != nil {
		log.Printf("Failed to serialize order placed event: %v", err)
		return
	}

	err = Channel.Publish(
		"",             // exchange
		"order_placed", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		log.Printf("Failed to publish order_placed event: %v", err)
	} else {
		log.Printf("Order Placed Event emitted: %s", body)
	}
}

func EmitOrderShipped(orderID int) {
	event := models.OrderShippedEvent{
		OrderID: orderID,
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to serialize order shipped event: %v", err)
		return
	}

	err = Channel.Publish(
		"",              // exchange
		"order_shipped", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		log.Printf("Failed to publish order_shipped event: %v", err)
	} else {
		log.Printf("Order Shipped Event emitted: %s", body)
	}
}

var ProductCatalog = make(map[int]models.Product)
var UserRegistry = make(map[int]models.User)
var mutex = &sync.RWMutex{}

func ListenForEvents() {
	// Listen for "Product Created" events
	go listenForProductCreated()

	// Listen for "User Registered" events
	go listenForUserRegistered()
}

func listenForProductCreated() {
	msgs, err := Channel.Consume(
		"product_created", // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer for product_created: %v", err)
	}

	go func() {
		for d := range msgs {
			var product models.Product
			err := json.Unmarshal(d.Body, &product)
			if err != nil {
				log.Printf("Failed to parse product created event: %v", err)
				continue
			}
			log.Printf("Received Product Created Event: %+v", product)
			mutex.Lock()
			ProductCatalog[product.ID] = product
			mutex.Unlock()
		}
	}()
}

func listenForUserRegistered() {
	msgs, err := Channel.Consume(
		"user_registered", // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer for user_registered: %v", err)
	}

	go func() {
		for d := range msgs {
			var user models.User
			err := json.Unmarshal(d.Body, &user)
			if err != nil {
				log.Printf("Failed to parse user registered event: %v", err)
				continue
			}
			log.Printf("Received User Registered Event: %+v", user)
			mutex.Lock()
			UserRegistry[user.ID] = user
			mutex.Unlock()
		}
	}()
}

func LoadExistingProducts() {
	url := utils.ProductServiceURL + "/products"
	resp, err := utils.HTTPClient.Get(url)
	if err != nil {
		log.Printf("Failed to fetch products from Product Service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to fetch products: %s", string(bodyBytes))
		return
	}

	var responseData map[string][]models.Product
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		log.Printf("Failed to parse products response: %v", err)
		return
	}

	products, exists := responseData["products"]
	if !exists {
		log.Printf("No 'products' field in response")
		return
	}

	mutex.Lock()
	for _, product := range products {
		ProductCatalog[product.ID] = product
	}
	mutex.Unlock()

	log.Printf("Loaded %d products from Product Service", len(products))
}
