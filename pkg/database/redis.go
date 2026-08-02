package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedis(env *viper.Viper, log *zap.Logger) (*redis.Client, error) {
	config := RedisConfig{
		Host:     env.GetString("REDIS_HOST"),
		Port:     env.GetString("REDIS_PORT"),
		Password: env.GetString("REDIS_PASSWORD"),
		DB:       env.GetInt("REDIS_DB"),
	}

	// Create redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Ping Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	log.Info("Redis connected successfully",
		zap.String("host", config.Host),
		zap.String("port", config.Port),
	)

	return client, nil
}
