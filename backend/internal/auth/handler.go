package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
}

// Request/response structs

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type userResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}

func newAuthResponse(t *Tokens) authResponse {
	return authResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		User: userResponse{
			ID:    t.UserID,
			Email: t.Email,
			Role:  string(t.Role),
		},
	}
}

// Handlers

// @Summary     Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Email and password"
// @Success     201 {object} authResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/register [post]
func (h *Handler) register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	tokens, err := h.svc.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrWeakPassword):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		case errors.Is(err, ErrEmailAlreadyTaken):
			return c.JSON(http.StatusConflict, echo.Map{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusCreated, newAuthResponse(tokens))
}

// @Summary     Login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Email and password"
// @Success     200 {object} authResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/login [post]
func (h *Handler) login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	tokens, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, newAuthResponse(tokens))
}

// @Summary     Refresh tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body refreshRequest true "Refresh token"
// @Success     200 {object} authResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/refresh [post]
func (h *Handler) refresh(c echo.Context) error {
	var req refreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	tokens, err := h.svc.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, newAuthResponse(tokens))
}

// @Summary     Logout
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body logoutRequest true "Refresh token"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/logout [post]
func (h *Handler) logout(c echo.Context) error {
	var req logoutRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	if err := h.svc.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.NoContent(http.StatusNoContent)
}
