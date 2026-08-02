package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxIdleConns int
	MaxOpenConns int
}

func NewPostgres(env *viper.Viper, log *zap.Logger) (*gorm.DB, error) {
	config := PostgresConfig{
		Host:         env.GetString("DB_HOST"),
		Port:         env.GetString("DB_PORT"),
		User:         env.GetString("DB_USER"),
		Password:     env.GetString("DB_PASSWORD"),
		DBName:       env.GetString("DB_NAME"),
		SSLMode:      env.GetString("DB_SSL_MODE"),
		MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS"),
		MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS"),
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	// Configure gorm logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if env.GetString("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	// Get underlying SQL Database
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
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Postgresql connected successfully", zap.String("host", config.Host), zap.String("database", config.DBName))

	return db, nil
}
