package routes

import (
	"test-lintaspay/internal/domain/entity"
	"test-lintaspay/internal/handler"
	"test-lintaspay/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

type Routes struct {
	authMW       *middlewares.AuthMiddleware
	health       *handler.HealthHandler
	auth         *handler.AuthHandler
	disbursement *handler.DisbursementHandler
	auditLog     *handler.AuditLogHandler
}

func NewRoutes(
	authMW *middlewares.AuthMiddleware,
	health *handler.HealthHandler,
	auth *handler.AuthHandler,
	disbursement *handler.DisbursementHandler,
	auditLog *handler.AuditLogHandler,
) *Routes {
	return &Routes{
		authMW:       authMW,
		health:       health,
		auth:         auth,
		disbursement: disbursement,
		auditLog:     auditLog,
	}
}

// Register maps all endpoints under the /api/v1 group.
func (r *Routes) Register(router fiber.Router) {
	router.Get("/health", r.health.Check)

	auth := router.Group("/auth")
	auth.Post("/login", r.auth.Login)
	auth.Post("/refresh", r.auth.Refresh)
	auth.Post("/logout", r.auth.Logout)

	generalLimit := middlewares.NewRateLimiter(120)
	createLimit := middlewares.NewRateLimiter(30)

	disbursement := router.Group("/disbursements", r.authMW.Protect(), generalLimit)
	disbursement.Get("/", r.disbursement.List)
	disbursement.Post("/", createLimit, r.disbursement.Create)
	disbursement.Post("/batch", createLimit, r.disbursement.CreateBatch)
	disbursement.Get("/:id", r.disbursement.Detail)
	disbursement.Patch("/:id/status",
		r.authMW.RequireRoles(entity.RoleAdmin, entity.RoleSuperadmin),
		r.disbursement.UpdateStatus,
	)
	disbursement.Delete("/:id",
		r.authMW.RequireRoles(entity.RoleSuperadmin),
		r.disbursement.Delete,
	)

	auditLog := router.Group("/audit-logs",
		r.authMW.Protect(),
		generalLimit,
		r.authMW.RequireRoles(entity.RoleSuperadmin),
	)
	auditLog.Get("/", r.auditLog.List)
}
