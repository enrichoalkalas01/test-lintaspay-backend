package routes

import (
	"test-lintaspay/internal/handler"
	"test-lintaspay/internal/repository"
	"test-lintaspay/internal/usecase"
	"test-lintaspay/pkg/middlewares"
	"test-lintaspay/pkg/server"

	"go.uber.org/fx"
)

var Module = fx.Module("internal",
	fx.Provide(
		// middlewares
		middlewares.NewAuthMiddleware,

		// health
		repository.NewHealthRepository,
		usecase.NewHealthUsecase,
		handler.NewHealthHandler,

		// auth
		repository.NewAuthRepository,
		usecase.NewAuthUsecase,
		handler.NewAuthHandler,

		// disbursements
		repository.NewDisbursementRepository,
		usecase.NewDisbursementUsecase,
		handler.NewDisbursementHandler,

		// audit logs
		repository.NewAuditLogRepository,
		usecase.NewAuditLogUsecase,
		handler.NewAuditLogHandler,

		// routes
		server.AsRoute(NewRoutes),
	),
)
