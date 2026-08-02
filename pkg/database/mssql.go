package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MSSQLConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	MaxIdleConns int
	MaxOpenConns int
}

func NewMSSQL(cfg *viper.Viper, log *zap.Logger) (*gorm.DB, error) {
	config := MSSQLConfig{
		Host:         cfg.GetString("MSSQL_HOST"),
		Port:         cfg.GetString("MSSQL_PORT"),
		User:         cfg.GetString("MSSQL_USER"),
		Password:     cfg.GetString("MSSQL_PASSWORD"),
		Database:     cfg.GetString("MSSQL_DATABASE"),
		MaxIdleConns: cfg.GetInt("MSSQL_MAX_IDLE_CONNS"),
		MaxOpenConns: cfg.GetInt("MSSQL_MAX_OPEN_CONNS"),
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.GetString("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open connection
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mssql: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mssql: %w", err)
	}

	log.Info("MSSQL connected successfully",
		zap.String("host", config.Host),
		zap.String("database", config.Database),
	)

	return db, nil
}
