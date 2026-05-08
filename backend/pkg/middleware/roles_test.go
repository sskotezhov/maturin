package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	mw "github.com/sskotezhov/maturin/pkg/middleware"
	"github.com/sskotezhov/maturin/pkg/roles"
)

func runRequireRoles(t *testing.T, contextRole string, allowed ...roles.Role) (code int, called bool) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if contextRole != "" {
		c.Set(mw.ContextRole, contextRole)
	}

	handler := mw.RequireRoles(allowed...)(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	})
	_ = handler(c)
	return rec.Code, called
}

func TestRequireRoles_MatchingSingleRole(t *testing.T) {
	code, called := runRequireRoles(t, "manager", roles.RoleManager)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, called)
}

func TestRequireRoles_MatchingOneOfMany(t *testing.T) {
	code, called := runRequireRoles(t, "admin", roles.RoleManager, roles.RoleAdmin)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, called)
}

func TestRequireRoles_WrongRole(t *testing.T) {
	code, called := runRequireRoles(t, "client", roles.RoleManager, roles.RoleAdmin)
	assert.Equal(t, http.StatusForbidden, code)
	assert.False(t, called)
}

func TestRequireRoles_NoRoleInContext(t *testing.T) {
	code, called := runRequireRoles(t, "", roles.RoleManager)
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.False(t, called)
}
