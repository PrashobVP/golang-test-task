package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"gopkg.in/redis.v5"
)

// Message structure matches the JSON format sent by the API
type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://user:password@localhost:7001/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	queueName := "messageprocessor"

	// Consume messages from the queue
	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to consume messages from queue:", err)
	}

	// Initialize Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	for msg := range msgs {
		// Deserialize the message body into the Message struct
		var receivedMsg Message
		err := json.Unmarshal(msg.Body, &receivedMsg)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Construct the Redis key for storing the message
		key := fmt.Sprintf("message:%s:%s", receivedMsg.Sender, receivedMsg.Receiver)

		// Save the message in Redis under the constructed key
		err = client.RPush(key, msg.Body).Err()
		if err != nil {
			log.Printf("Error saving message to Redis: %v", err)
		}

		fmt.Printf("Received and saved message: %s\n", msg.Body)
	}
}
