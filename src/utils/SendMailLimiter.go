package utils

import (
	"auth_service/logger"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SendMailLimiter membatasi pengiriman email berdasarkan Redis cache
func SendMailLimiter(redisClient *redis.Client, referenceID, email, event string, waitingTime time.Duration) (time.Duration, error) {
	if redisClient == nil {
		logger.Error(referenceID, "ERROR - SendMailLimiter - Redis client is not initialized")
		return 0, fmt.Errorf("internal server error: Redis client is not initialized")
	}

	redisKey := generateRedisKey(event, email)
	logger.Info(referenceID, fmt.Sprintf("INFO - SendMailLimiter - Checking rate limit for event: %s, email: %s, key: %s, duration: %v", event, email, redisKey, waitingTime))

	// Cek apakah sudah ada request sebelumnya
	exists, err := redisClient.Exists(context.Background(), redisKey).Result()
	if err != nil {
		logger.Error(referenceID, "ERROR - SendMailLimiter - Failed to check Redis key existence: ", err)
		return 0, fmt.Errorf("internal server error")
	}

	logger.Info(referenceID, fmt.Sprintf("INFO - SendMailLimiter - Key existence check result: %d", exists))

	// Jika request sudah ada, ambil waktu sisa (TTL)
	if exists > 0 {
		ttl, err := redisClient.TTL(context.Background(), redisKey).Result()
		if err != nil {
			logger.Error(referenceID, "ERROR - SendMailLimiter - Failed to get TTL from Redis: ", err)
			return 0, fmt.Errorf("internal server error")
		}
		logger.Error(referenceID, fmt.Sprintf("ERROR - SendMailLimiter - Too many %s requests for email: %s, TTL remaining: %v", event, email, ttl))
		return ttl, fmt.Errorf("too many %s requests", event)
	}

	// Simpan request dalam Redis dengan expiry time sesuai waitingTime yang diberikan
	setResult, err := redisClient.SetEx(context.Background(), redisKey, "requested", waitingTime).Result()
	if err != nil {
		logger.Error(referenceID, fmt.Sprintf("ERROR - SendMailLimiter - Failed to store %s request in Redis: ", event), err)
		return 0, fmt.Errorf("internal server error")
	}

	logger.Info(referenceID, fmt.Sprintf("INFO - SendMailLimiter - Key successfully set in Redis: %s, Result: %s", redisKey, setResult))

	return 0, nil
}

// generateRedisKey membuat key unik untuk setiap event dan email
func generateRedisKey(event, email string) string {
	return fmt.Sprintf("email_limit:%s:%s", event, email)
}
