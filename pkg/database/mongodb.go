package database

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoConfig struct {
	URI      string
	Database string
}

func NewMongoDB(env *viper.Viper, log *zap.Logger) (*mongo.Database, error) {
	config := MongoConfig{
		URI:      env.GetString("MONGO_URI"),
		Database: env.GetString("MONGO_DATABASE"),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to mongodb
	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect mongodb: %w", err)
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	log.Info("MongoDB connected successfully",
		zap.String("database", config.Database),
	)

	return client.Database(config.Database), nil
}
