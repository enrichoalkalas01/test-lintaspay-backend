package configs

import (
	"test-lintaspay/pkg/database"
	"test-lintaspay/pkg/logger"
	"test-lintaspay/pkg/server"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Config Vipers
func NewViperFromEnv() (*viper.Viper, error) {
	return NewViper(".env", "env", ".", "../../")
}

var ConfigModule = fx.Module("config", fx.Provide(NewViperFromEnv))

// Logger
var LoggerModule = fx.Module("logger", fx.Provide(logger.NewZapLogger))

// Database
var DatabaseModule = fx.Module("database", fx.Provide(database.NewMySQL))

// Server
var ServerModule = fx.Module("server",
	fx.Provide(server.NewFiberApp),
	fx.Invoke(server.RegisterHooks),
)
