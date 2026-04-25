package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/sskotezhov/maturin/pkg/roles"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

type accessClaims struct {
	UserID uint   `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing or invalid authorization header"})
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if tokenStr == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing or invalid authorization header"})
			}

			claims := &accessClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid or expired token"})
			}

			if claims.UserID == 0 {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
			}
			if !roles.Valid(claims.Role) {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
			}

			c.Set(ContextUserID, claims.UserID)
			c.Set(ContextRole, claims.Role)

			return next(c)
		}
	}
}
