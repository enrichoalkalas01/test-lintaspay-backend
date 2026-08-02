package main

import (
	"test-lintaspay/configs"
	_ "test-lintaspay/docs"
	"test-lintaspay/internal/routes"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		configs.ConfigModule,
		configs.LoggerModule,
		configs.DatabaseModule,
		configs.ServerModule,

		routes.Module,
	).Run()
}
