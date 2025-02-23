package utils

import (
	"auth_service/logger"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func OTPRateLimiter(redisClient *redis.Client, email string) ( time.Duration, error) {
	if redisClient == nil {
		return 0, fmt.Errorf("internal server error: Redis client is not initialized")
	}

	redisKey := fmt.Sprintf("otp_request:%s", email)

	// Cek apakah ada request OTP dalam waktu 60 detik terakhir
	exists, err := redisClient.Exists(context.Background(), redisKey).Result()
	if err != nil {
		logger.Error("ERROR - OTP RateLimiter - Failed to check Redis: ", err)
		return 0,  fmt.Errorf("internal server error")
	}

	// Jika request OTP sudah ada, ambil waktu sisa (TTL)
	if exists > 0 {
		ttl, err := redisClient.TTL(context.Background(), redisKey).Result()
		if err != nil {
			logger.Error("ERROR - OTP RateLimiter - Failed to get TTL from Redis: ", err)
			return 0,  fmt.Errorf("internal server error")
		}

		return ttl, fmt.Errorf("too many OTP requests")
	}

	// Simpan email dalam Redis dengan expiry 60 detik (1 menit)
	err = redisClient.Set(context.Background(), redisKey, "requested", 60*time.Second).Err()
	if err != nil {
		logger.Error("ERROR - OTP RateLimiter - Failed to store OTP request in Redis: ", err)
		return 0, fmt.Errorf("internal server error")
	}

	return 0, nil
}
