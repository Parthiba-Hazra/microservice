package rabbitmq

import (
	"encoding/json"
	"log"
	"user_service/models"

	"github.com/streadway/amqp"
)

func EmitUserRegistered(user models.User) {
	body, err := json.Marshal(user)
	if err != nil {
		log.Printf("Failed to serialize user: %v", err)
		return
	}

	err = Channel.Publish(
		"",                // exchange
		"user_registered", // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		log.Printf("Failed to publish user_registered event: %v", err)
	} else {
		log.Printf("User Registered Event emitted: %s", body)
	}
}

func EmitUserProfileUpdated(user models.User) {
	body, err := json.Marshal(user)
	if err != nil {
		log.Printf("Failed to serialize user: %v", err)
		return
	}

	err = Channel.Publish(
		"",                     // exchange
		"user_profile_updated", // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		log.Printf("Failed to publish user_profile_updated event: %v", err)
	} else {
		log.Printf("User Profile Updated Event emitted: %s", body)
	}
}
