package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type TokenRepository struct {
	Client *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepository {
	return &TokenRepository{Client: client}
}

// StoreToken stores a token in Redis with an expiration time.
func (repo *TokenRepository) StoreToken(ctx context.Context, token string, userID int, expiration time.Duration) error {
	key := repo.buildKey(token, userID)
	return repo.Client.Set(ctx, key, "valid", expiration).Err()
}

// IsTokenValid checks if a token is valid (exists in Redis).
func (repo *TokenRepository) IsTokenValid(ctx context.Context, token string, userID int) (bool, error) {
	key := repo.buildKey(token, userID)
	exists, err := repo.Client.Exists(ctx, key).Result()
	return exists == 1, err
}

// DeleteToken removes a token from Redis.
func (repo *TokenRepository) DeleteToken(ctx context.Context, token string, userID int) error {
	key := repo.buildKey(token, userID)
	return repo.Client.Del(ctx, key).Err()
}

// buildKey generates a unique Redis key for the token.
func (repo *TokenRepository) buildKey(token string, userID int) string {
	return fmt.Sprintf("user:%d:token:%s", userID, token)
}
