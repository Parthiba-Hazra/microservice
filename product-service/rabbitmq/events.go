package rabbitmq

import (
	"encoding/json"
	"log"
	"product-service/db"
	"product-service/models"

	"github.com/streadway/amqp"
)

func EmitProductCreated(product models.Product) {
	body, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to serialize product: %v", err)
		return
	}

	err = Channel.Publish(
		"",                // exchange
		"product_created", // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		log.Printf("Failed to publish product_created event: %v", err)
	} else {
		log.Printf("Product Created Event emitted: %s", body)
	}
}

func EmitInventoryUpdated(productID int, newInventory int) {
	event := models.InventoryUpdateEvent{
		ProductID:    productID,
		NewInventory: newInventory,
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to serialize inventory update event: %v", err)
		return
	}

	err = Channel.Publish(
		"",                  // exchange
		"inventory_updated", // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		log.Printf("Failed to publish inventory_updated event: %v", err)
	} else {
		log.Printf("Inventory Updated Event emitted: %s", body)
	}
}

func EmitProductUpdated(product models.Product) {
	body, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to serialize product: %v", err)
		return
	}

	err = Channel.Publish(
		"",                // exchange
		"product_updated", // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		log.Printf("Failed to publish product_updated event: %v", err)
	} else {
		log.Printf("Product Updated Event emitted: %s", body)
	}
}

func EmitProductDeleted(productID int) {
	event := map[string]int{
		"product_id": productID,
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to serialize product deleted event: %v", err)
		return
	}

	err = Channel.Publish(
		"",                // exchange
		"product_deleted", // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		log.Printf("Failed to publish product_deleted event: %v", err)
	} else {
		log.Printf("Product Deleted Event emitted: %s", body)
	}
}

func ListenForOrderPlacedEvents() {
	msgs, err := Channel.Consume(
		"order_placed", // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			var orderEvent models.OrderPlacedEvent
			err := json.Unmarshal(d.Body, &orderEvent)
			if err != nil {
				log.Printf("Failed to parse order placed event: %v", err)
				continue
			}
			log.Printf("Received Order Placed Event: %+v", orderEvent)
			// Update inventory based on the order
			updateInventory(orderEvent)
		}
	}()
}

func updateInventory(orderEvent models.OrderPlacedEvent) {
	for _, item := range orderEvent.Items {
		// Modify the query to return the updated inventory after the update
		query := `UPDATE products SET inventory = inventory - $1 WHERE id = $2 RETURNING inventory`

		var updatedInventory int
		err := db.DB.QueryRow(query, item.Quantity, item.ProductID).Scan(&updatedInventory)
		if err != nil {
			log.Printf("Failed to update inventory for product %d: %v", item.ProductID, err)
		} else {
			EmitInventoryUpdated(item.ProductID, updatedInventory)
			log.Printf("Inventory updated for product %d, new inventory: %d", item.ProductID, updatedInventory)
		}
	}
}
