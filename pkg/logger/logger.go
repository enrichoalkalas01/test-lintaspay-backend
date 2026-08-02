package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

type Config struct {
	Environtment string // "development" or "production"
	LogLevel     string // "debug", "info", "warn", "error"
	OutputPath   string // "stdout" or file path
	Format       string // "json" or "console", overrides environment default when set
}

// ===== Init Function =====
func NewLogger(config Config) (*Logger, error) {
	var zapConfig zap.Config

	// Set base config based on environment
	if config.Environtment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Encoding = "json"
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// allow forcing the output format regardless of environment
	if config.Format == "json" || config.Format == "console" {
		zapConfig.Encoding = config.Format
		if config.Format == "json" {
			zapConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		}
	}

	// Set Log Level
	level := zapcore.InfoLevel
	switch config.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Configure time format
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Set Output Path
	if config.OutputPath == "" || config.OutputPath == "stdout" {
		zapConfig.OutputPaths = []string{"stdout"}
	} else {
		zapConfig.OutputPaths = []string{config.OutputPath}
	}

	zapConfig.ErrorOutputPaths = []string{"stderr"}

	// Build Logger
	log, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	if err != nil {
		return nil, err
	}

	return &Logger{Logger: log}, nil
}

// ===== Helper Functions =====
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.Logger.Panic(msg, fields...)
}

// ===== WithFields returns a logger with pre-set fields =====
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// ===== Sync flushes any buffered log entries =====
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger() (*Logger, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return NewLogger(Config{
		Environtment: env,
		LogLevel:     logLevel,
		OutputPath:   "stdout",
		Format:       os.Getenv("LOG_FORMAT"),
	})
}

func NewZapLogger() (*zap.Logger, error) {
	logger, err := NewDefaultLogger()
	if err != nil {
		return nil, err
	}

	return logger.Logger, nil
}

/*

// Example usage dengan fields
	log.Info("Server configuration",
		zap.String("host", "localhost"),
		zap.Int("port", 8080),
		zap.String("version", "1.0.0"),
	)

	// Example with structured fields
	log.WithFields(
		zap.String("module", "database"),
		zap.String("driver", "postgres"),
	).Info("Database connection established")

	// Example error logging
	log.Error("Example error log",
		zap.String("error_code", "ERR_001"),
		zap.String("description", "This is an example error"),
	)
*/
