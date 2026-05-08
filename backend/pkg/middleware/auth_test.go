package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	mw "github.com/sskotezhov/maturin/pkg/middleware"
)

const testJWTSecret = "test-secret"

func signToken(userID uint, role string, secret string, ttl time.Duration) string {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(ttl).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := t.SignedString([]byte(secret))
	return s
}

func runJWT(t *testing.T, authHeader string) (code int, uid any, role any, called bool) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw.JWTAuth(testJWTSecret)(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	})
	_ = handler(c)
	return rec.Code, c.Get(mw.ContextUserID), c.Get(mw.ContextRole), called
}

func TestJWTAuth_ValidToken(t *testing.T) {
	token := signToken(42, "client", testJWTSecret, time.Hour)
	code, uid, role, called := runJWT(t, "Bearer "+token)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, called)
	assert.Equal(t, uint(42), uid)
	assert.Equal(t, "client", role)
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	code, _, _, called := runJWT(t, "")
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}

func TestJWTAuth_NonBearerScheme(t *testing.T) {
	token := signToken(1, "client", testJWTSecret, time.Hour)
	code, _, _, called := runJWT(t, "Basic "+token)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	token := signToken(1, "client", "wrong-secret", time.Hour)
	code, _, _, called := runJWT(t, "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	token := signToken(1, "client", testJWTSecret, -time.Hour)
	code, _, _, called := runJWT(t, "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}

func TestJWTAuth_ZeroUserID(t *testing.T) {
	token := signToken(0, "client", testJWTSecret, time.Hour)
	code, _, _, called := runJWT(t, "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}

func TestJWTAuth_InvalidRole(t *testing.T) {
	token := signToken(1, "superadmin", testJWTSecret, time.Hour)
	code, _, _, called := runJWT(t, "Bearer "+token)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}
