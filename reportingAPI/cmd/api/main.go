package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sort"
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Message struct {
    Sender   string `json:"sender"`
    Receiver string `json:"receiver"`
    Content  string `json:"message"`
}

func main() {
    r := gin.Default()

    redisClient := redis.NewClient(&redis.Options{
        Addr: "redis:6379",
        Password: "",
        DB: 0,
    })

    r.GET("/message/list", func(c *gin.Context) {
        sender := c.Query("sender")
        receiver := c.Query("receiver")

        if sender == "" || receiver == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "sender and receiver are required"})
            return
        }

        key := sender + ":" + receiver
        messages, err := redisClient.LRange(ctx, key, 0, -1).Result()
        if err != nil {
            log.Printf("Failed to retrieve messages from Redis: %s", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve messages"})
            return
        }

        var result []Message
        for _, msgStr := range messages {
            var msg Message
            if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
                log.Printf("Failed to unmarshal message: %s", err)
                continue
            }
            result = append(result, msg)
        }

        // Sort messages in reverse order 
        sort.Slice(result, func(i, j int) bool {
            return i > j
        })

        c.JSON(http.StatusOK, result)
    })

    r.Run(":8085")
}
