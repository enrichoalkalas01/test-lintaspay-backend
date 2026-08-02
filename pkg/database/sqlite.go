package database

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteConfig struct {
	Path string
}

func NewSQLite(cfg *viper.Viper, log *zap.Logger) (*gorm.DB, error) {
	config := SQLiteConfig{
		Path: cfg.GetString("SQLITE_PATH"),
	}

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.GetString("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open connection
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	log.Info("SQLite connected successfully",
		zap.String("path", config.Path),
	)

	return db, nil
}
