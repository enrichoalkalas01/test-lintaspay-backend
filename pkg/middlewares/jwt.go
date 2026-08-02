package middlewares

import (
	"slices"
	"strings"

	"test-lintaspay/pkg/common/httpresponse"
	"test-lintaspay/pkg/common/reqctx"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type AuthMiddleware struct {
	secret []byte
}

func NewAuthMiddleware(env *viper.Viper) *AuthMiddleware {
	return &AuthMiddleware{
		secret: []byte(env.GetString("JWT_SECRET")),
	}
}

// Protect validates the bearer token and stores the claims into locals
// and the user context for downstream layers.
func (m *AuthMiddleware) Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return httpresponse.Unauthorized("missing or malformed authorization header")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.secret, nil
		})
		if err != nil || !token.Valid {
			return httpresponse.Unauthorized("invalid or expired token")
		}

		userID, _ := claims["sub"].(string)
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		c.Locals("user_id", userID)
		c.Locals("username", username)
		c.Locals("role", role)
		c.SetUserContext(reqctx.WithUser(c.UserContext(), userID, username, role))

		return c.Next()
	}
}

// RequireRoles rejects the request when the authenticated role is not listed.
func (m *AuthMiddleware) RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)

		if slices.Contains(roles, role) {
			return c.Next()
		}

		return httpresponse.Forbidden("you are not allowed to access this resource")
	}
}
