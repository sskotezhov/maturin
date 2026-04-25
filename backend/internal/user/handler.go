package user

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	mw "github.com/sskotezhov/maturin/pkg/middleware"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/me", h.getProfile)
	g.PATCH("/me", h.updateProfile)
}

type profileResponse struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	LastName    string `json:"last_name"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name"`
	Phone       string `json:"phone"`
	Telegram    string `json:"telegram"`
	CompanyName string `json:"company_name"`
	INN         string `json:"inn"`
}

type updateProfileRequest struct {
	Mask   []string          `json:"mask"`
	Values map[string]string `json:"values"`
}

func toProfileResponse(u *User) profileResponse {
	return profileResponse{
		ID:          u.ID,
		Email:       u.Email,
		Role:        string(u.Role),
		LastName:    u.LastName,
		FirstName:   u.FirstName,
		MiddleName:  u.MiddleName,
		Phone:       u.Phone,
		Telegram:    u.Telegram,
		CompanyName: u.CompanyName,
		INN:         u.INN,
	}
}

// @Summary     Get current user profile
// @Tags        user
// @Produce     json
// @Success     200 {object} profileResponse
// @Failure     401 {object} map[string]string
// @Security    BearerAuth
// @Router      /user/me [get]
func (h *Handler) getProfile(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)

	u, err := h.svc.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	return c.JSON(http.StatusOK, toProfileResponse(u))
}

// @Summary     Update current user profile
// @Tags        user
// @Accept      json
// @Produce     json
// @Param       body body updateProfileRequest true "Mask and values"
// @Success     200 {object} profileResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Security    BearerAuth
// @Router      /user/me [patch]
func (h *Handler) updateProfile(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)

	var req updateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if len(req.Mask) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "mask is empty"})
	}

	u, err := h.svc.UpdateProfile(c.Request().Context(), userID, UpdateInput{
		Mask:   req.Mask,
		Values: req.Values,
	})
	if err != nil {
		if errors.Is(err, ErrForbiddenField) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, toProfileResponse(u))
}
