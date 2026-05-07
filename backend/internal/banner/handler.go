package banner

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sskotezhov/maturin/internal/product"
)

type Handler struct {
	svc        Service
	productSvc product.Service
}

func NewHandler(svc Service, productSvc product.Service) *Handler {
	return &Handler{svc: svc, productSvc: productSvc}
}

// Register registers the public GET route.
func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.get)
}

// RegisterAdmin registers the admin-only PUT route.
func (h *Handler) RegisterAdmin(g *echo.Group) {
	g.PUT("", h.set)
}

type bannerItemResponse struct {
	Position  int              `json:"position"`
	ProductID string           `json:"product_id"`
	Product   *product.Product `json:"product"`
}

type getBannerResponse struct {
	Items []bannerItemResponse `json:"items"`
}

// @Summary     Get banner products
// @Tags        banner
// @Produce     json
// @Success     200 {object} getBannerResponse
// @Failure     500 {object} map[string]string
// @Router      /banner [get]
func (h *Handler) get(c echo.Context) error {
	items, err := h.svc.Get(c.Request().Context())
	if err != nil {
		slog.Error("banner get failed", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	resp := make([]bannerItemResponse, 0, len(items))
	for _, item := range items {
		p, err := h.productSvc.GetProduct(c.Request().Context(), item.ProductID)
		if err != nil {
			// skip products no longer in 1C
			resp = append(resp, bannerItemResponse{
				Position:  item.Position,
				ProductID: item.ProductID,
				Product:   nil,
			})
			continue
		}
		resp = append(resp, bannerItemResponse{
			Position:  item.Position,
			ProductID: item.ProductID,
			Product:   p,
		})
	}

	return c.JSON(http.StatusOK, getBannerResponse{Items: resp})
}

type setBannerRequest struct {
	ProductIDs []string `json:"product_ids"`
}

// @Summary     Set banner products (admin only)
// @Tags        banner
// @Accept      json
// @Produce     json
// @Param       body body setBannerRequest true "Exactly 4 product IDs"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Security    BearerAuth
// @Router      /admin/banner [put]
func (h *Handler) set(c echo.Context) error {
	var req setBannerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	if err := h.svc.Set(c.Request().Context(), req.ProductIDs); err != nil {
		switch {
		case errors.Is(err, ErrInvalidCount):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "banner must contain exactly 4 products"})
		case errors.Is(err, ErrDuplicateProduct):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "product IDs must be unique"})
		default:
			slog.Error("banner set failed", "err", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}
