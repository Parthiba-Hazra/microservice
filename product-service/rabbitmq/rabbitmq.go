package rabbitmq

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

func Init() {
	var err error
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	maxRetries := 12
	for retries := 0; retries < maxRetries; retries++ {
		Conn, err = amqp.Dial(rabbitmqURL)
		if err != nil {
			log.Printf("Failed to connect to RabbitMQ: %v", err)
		} else {
			Channel, err = Conn.Channel()
			if err == nil {
				log.Println("Successfully connected to RabbitMQ")
				break
			}
			log.Printf("Failed to open a channel: %v", err)
		}
		time.Sleep(5 * time.Second)
		log.Println("Retrying RabbitMQ connection...")
	}

	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ after %d attempts", maxRetries)
	}

	declareQueues()
}

func declareQueues() {
	queues := []string{"product_created", "product_updated", "product_deleted", "inventory_updated", "order_placed"}

	for _, queueName := range queues {
		_, err := Channel.QueueDeclare(
			queueName,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", queueName, err)
		}
	}
}

func Close() {
	Channel.Close()
	Conn.Close()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
