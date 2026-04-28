package inquiry

import (
	"errors"
	"log/slog"
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
	g.POST("", h.submit)
}

type submitRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Comment string `json:"comment"`
	Consent bool   `json:"consent"`
	Source  string `json:"source"`
	PageURL string `json:"page_url"`
}

type submitResponse struct {
	ID      uint   `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// @Summary     Submit public inquiry form
// @Tags        inquiries
// @Accept      json
// @Produce     json
// @Param       body body submitRequest true "Inquiry form"
// @Success     201 {object} submitResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /inquiries [post]
func (h *Handler) submit(c echo.Context) error {
	var req submitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	item, err := h.svc.Submit(c.Request().Context(), SubmitInput{
		Name:    req.Name,
		Phone:   req.Phone,
		Comment: req.Comment,
		Consent: req.Consent,
		Source:  req.Source,
		PageURL: req.PageURL,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrConsentRequired):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "consent is required"})
		case errors.Is(err, ErrInvalidName):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid name"})
		case errors.Is(err, ErrInvalidPhone):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid phone"})
		case errors.Is(err, ErrCommentTooLong):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "comment is too long"})
		default:
			slog.Error("inquiry submit failed", "err", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusCreated, submitResponse{
		ID:      item.ID,
		Status:  string(item.Status),
		Message: "inquiry submitted",
	})
}
