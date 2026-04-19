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
	g.POST("/verify-email", h.verifyEmail)
	g.POST("/resend-verification", h.resendVerification)
	g.POST("/login", h.login)
	g.POST("/refresh", h.refresh)
	g.POST("/logout", h.logout)
	g.POST("/forgot-password", h.forgotPassword)
	g.POST("/reset-password", h.resetPassword)
}

// Request/response structs

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
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
// @Success     201 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/register [post]
func (h *Handler) register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	err := h.svc.Register(c.Request().Context(), req.Email, req.Password)
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

	return c.JSON(http.StatusCreated, echo.Map{"message": "verification code sent to email"})
}

// @Summary     Verify email with code
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body verifyEmailRequest true "Email and verification code"
// @Success     200 {object} authResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/verify-email [post]
func (h *Handler) verifyEmail(c echo.Context) error {
	var req verifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	tokens, err := h.svc.VerifyEmail(c.Request().Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCode):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		case errors.Is(err, ErrInvalidCredentials):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusOK, newAuthResponse(tokens))
}

// @Summary     Resend email verification code
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body resendVerificationRequest true "Email"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/resend-verification [post]
func (h *Handler) resendVerification(c echo.Context) error {
	var req resendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	if err := h.svc.ResendVerification(c.Request().Context(), req.Email); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "verification code sent to email"})
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
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		case errors.Is(err, ErrEmailNotVerified):
			return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
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

// @Summary     Send password reset code
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body forgotPasswordRequest true "Email"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/forgot-password [post]
func (h *Handler) forgotPassword(c echo.Context) error {
	var req forgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	if err := h.svc.ForgotPassword(c.Request().Context(), req.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "if this email exists, a reset code has been sent"})
}

// @Summary     Reset password with code
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body resetPasswordRequest true "Email, code and new password"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /auth/reset-password [post]
func (h *Handler) resetPassword(c echo.Context) error {
	var req resetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	err := h.svc.ResetPassword(c.Request().Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCode):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		case errors.Is(err, ErrWeakPassword):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "password has been reset"})
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
