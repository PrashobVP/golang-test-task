package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

func main() {
	r := gin.Default()

	r.POST("/message", func(c *gin.Context) {
		var requestBody struct {
			Sender   string `json:"sender"`
			Receiver string `json:"receiver"`
			Message  string `json:"message"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
			return
		}

		// Connect to RabbitMQ
		conn, err := amqp.Dial("amqp://user:password@localhost:7001/")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to RabbitMQ"})
			return
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open a channel"})
			return
		}
		defer ch.Close()

		queueName := "messageprocessor" 

		// Declare the queue 
		_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare queue"})
			return
		}

		// Publish the message to the queue
		err = ch.Publish("", queueName, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprintf("%s: %s", requestBody.Sender, requestBody.Message)),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.Run(":8080")
}
