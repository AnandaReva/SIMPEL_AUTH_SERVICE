package rds

import (
	"auth_service/logger"
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

/* !!NOTE : Ther should be only one redis client */ 

var (
	RedisClient *redis.Client
	redisMu     sync.Mutex
)
// InitRedisConn menginisialisasi Redis client
func InitRedisConn(host, pass string, db int) error {
	redisMu.Lock()
	defer redisMu.Unlock()

	if RedisClient != nil {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: pass,
		DB:       db,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		logger.Error("REDIS", fmt.Sprintf("ERROR - Redis connection failed: %v", err))
		client.Close()
		return err
	}

	RedisClient = client
	logger.Info("REDIS", "INFO - Successfully connected to Redis")
	return nil
}

// GetRedisClient memastikan Redis client aktif
func GetRedisClient() *redis.Client {
	redisMu.Lock()
	defer redisMu.Unlock()

	if RedisClient == nil {
		logger.Error("REDIS", "ERROR - Redis client is not initialized")
		return nil
	}

	if _, err := RedisClient.Ping(context.Background()).Result(); err != nil {
		logger.Error("REDIS", "ERROR - Redis connection lost. Reconnecting...")
		RedisClient.Close()
		RedisClient = nil
	}

	return RedisClient
}



