package server

import (
	"context"

	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Route is implemented by every handler that wants to register endpoints.
type Route interface {
	Register(router fiber.Router)
}

// AsRoute annotates a handler constructor so fx collects it into the "routes" group.
func AsRoute(constructor any) any {
	return fx.Annotate(
		constructor,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}

type Params struct {
	fx.In

	Env    *viper.Viper
	Log    *zap.Logger
	Routes []Route `group:"routes"`
}

func NewFiberApp(p Params) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      p.Env.GetString("APP_NAME"),
		ErrorHandler: httpresponse.NewErrorHandler(p.Log),
	})

	app.Use(middlewares.NewRequestIDMiddleware().Handle())
	app.Use(middlewares.NewRequestLoggerMiddleware(p.Log))
	app.Use(middlewares.NewRecoveryMiddleware(p.Log))

	if p.Env.GetString("APP_ENV") != "production" {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	api := app.Group("/api/v1")
	for _, route := range p.Routes {
		route.Register(api)
	}

	return app
}

func RegisterHooks(lc fx.Lifecycle, app *fiber.App, env *viper.Viper, log *zap.Logger) {
	port := env.GetString("SERVER_PORT")
	if port == "" {
		port = "3000"
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := app.Listen(":" + port); err != nil {
					log.Fatal("failed to start server", zap.Error(err))
				}
			}()
			log.Info("server started", zap.String("port", port))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down server")
			return app.ShutdownWithContext(ctx)
		},
	})
}
