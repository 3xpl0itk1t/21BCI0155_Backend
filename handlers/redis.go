package handlers

import (
	"context"
	"fmt"
	"log"
	"trademarkia/config"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDR, // Use container name and port
		Password: "",                // No password set
		DB:       0,                 // Default DB
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis!")
}
