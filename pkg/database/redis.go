package database

import (
	"context"
	"go-corenglish/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func ConnectRedis(cfg *config.Config, log *logrus.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.RedisPassword,
		DB:           0,
		PoolSize:     10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Errorf("Failed to connect to Redis: %v", err)
		return nil
	}

	log.Info("Successfully connected to Redis")
	return rdb
}
