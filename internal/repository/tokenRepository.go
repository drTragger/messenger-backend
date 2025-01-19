package repository

import (
	"context"
	"errors"
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
func (tr *TokenRepository) StoreToken(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	key := fmt.Sprintf("user:%d:token:%s", userID, token)
	return tr.Client.Set(ctx, key, "valid", expiration).Err()
}

// IsTokenValid checks if a token is valid (exists in Redis).
func (tr *TokenRepository) IsTokenValid(ctx context.Context, token string, userID uint) (bool, error) {
	key := fmt.Sprintf("user:%d:token:%s", userID, token)
	exists, err := tr.Client.Exists(ctx, key).Result()
	return exists == 1, err
}

// DeleteToken removes a token from Redis.
func (tr *TokenRepository) DeleteToken(ctx context.Context, token string, userID uint) error {
	key := fmt.Sprintf("user:%d:token:%s", userID, token)
	return tr.Client.Del(ctx, key).Err()
}

// StoreVerificationCode stores the verification code for a phone number to Redis
func (tr *TokenRepository) StoreVerificationCode(ctx context.Context, phone string, code string, expiry time.Duration) error {
	key := fmt.Sprintf("verification:%s", phone)
	return tr.Client.Set(ctx, key, code, expiry).Err()
}

// GetVerificationCode retrieves the verification code for a phone number from Redis
func (tr *TokenRepository) GetVerificationCode(ctx context.Context, phone string) (string, error) {
	key := fmt.Sprintf("verification:%s", phone)
	code, err := tr.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		// Key does not exist
		return "", fmt.Errorf("no verification code found for phone: %s", phone)
	}
	if err != nil {
		// Other Redis error
		return "", err
	}
	return code, nil
}

// DeleteVerificationCode deletes the verification code for a phone number from Redis
func (tr *TokenRepository) DeleteVerificationCode(ctx context.Context, phone string) error {
	key := fmt.Sprintf("verification:%s", phone)
	_, err := tr.Client.Del(ctx, key).Result()
	return err
}

// StoreResendCodeAttempt stores the phone of a user who resends the phone verification code
func (tr *TokenRepository) StoreResendCodeAttempt(ctx context.Context, phone string, expiry time.Duration) error {
	key := fmt.Sprintf("resendPhone:%s", phone)
	return tr.Client.Set(ctx, key, phone, expiry).Err()
}

// GetResendCodeAttempt retrieves the phone of a user who resends the phone verification code
func (tr *TokenRepository) GetResendCodeAttempt(ctx context.Context, phone string) (string, error) {
	key := fmt.Sprintf("resendPhone:%s", phone)
	code, err := tr.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// DeleteResendCodeAttempt deletes the phone of a user who resends the phone verification code
func (tr *TokenRepository) DeleteResendCodeAttempt(ctx context.Context, phone string) error {
	key := fmt.Sprintf("resendPhone:%s", phone)
	_, err := tr.Client.Del(ctx, key).Result()
	return err
}
