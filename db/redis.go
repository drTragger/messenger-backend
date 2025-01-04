package db

import (
	"github.com/go-redis/redis/v8"
	"log"
)

func InitRedis(redisAddr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr, // e.g., "localhost:6379"
	})

	// Ping Redis to ensure the connection is established
	if err := client.Ping(client.Context()).Err(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to Redis")
	return client, nil
}
