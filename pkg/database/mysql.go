package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	// MaxIdleConns int
	// MaxOpenConns int
	SSLMode string
	// SSLCA        string
}

func NewMySQL(env *viper.Viper, log *zap.Logger) (*gorm.DB, error) {
	config := MySQLConfig{
		Host:     env.GetString("MYSQL_HOST"),
		Port:     env.GetString("MYSQL_PORT"),
		User:     env.GetString("MYSQL_USER"),
		Password: env.GetString("MYSQL_PASSWORD"),
		Database: env.GetString("MYSQL_DATABASE"),
		// MaxIdleConns: env.GetInt("MYSQL_MAX_IDLE_CONNS"),
		// MaxOpenConns: env.GetInt("MYSQL_MAX_OPEN_CONNS"),
		SSLMode: env.GetString("MYSQL_SSL_MODE"),
		// SSLCA:        env.GetString("MYSQL_SSL_CA"),
	}

	// Build DSN
	// dsn := fmt.Sprintf(
	// 	"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	// 	config.User,
	// 	config.Password,
	// 	config.Host,
	// 	config.Port,
	// 	config.Database,
	// )

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if env.GetString("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool
	// sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	// sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	log.Info("MySQL connected successfully",
		zap.String("host", config.Host),
		zap.String("database", config.Database),
		zap.String("ssl_mode", config.SSLMode),
	)

	return db, nil
}
