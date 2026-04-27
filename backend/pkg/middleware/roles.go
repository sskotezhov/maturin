package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sskotezhov/maturin/pkg/roles"
)

func RequireRoles(allowed ...roles.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleStr, ok := c.Get(ContextRole).(string)
			if !ok || roleStr == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthenticated"})
			}
			for _, r := range allowed {
				if string(r) == roleStr {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, echo.Map{"error": "insufficient role"})
		}
	}
}
