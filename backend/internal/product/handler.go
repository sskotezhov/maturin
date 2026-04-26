package product

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("", h.listProducts)
	g.GET("/:id", h.getProduct)
}

func (h *Handler) RegisterCategories(g *echo.Group) {
	g.GET("", h.listCategories)
}

// @Summary     List products
// @Tags        products
// @Produce     json
// @Param       q         query string  false "Поиск по названию, артикулу, коду"
// @Param       category  query string  false "GUID категории"
// @Param       type      query string  false "Тип: Запас | Услуга"
// @Param       in_stock  query boolean false "Только в наличии"
// @Param       has_price query boolean false "Только с ценой"
// @Param       min_price query number  false "Минимальная цена"
// @Param       max_price query number  false "Максимальная цена"
// @Param       sort      query string  false "Сортировка: name | price | code | updated_at (default: name)"
// @Param       sort_dir  query string  false "Направление: asc | desc (default: asc)"
// @Param       page      query int     false "Страница (default: 1)"
// @Param       limit     query int     false "Кол-во на странице (default: 20)"
// @Success     200 {object} ListResult
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /products [get]
func (h *Handler) listProducts(c echo.Context) error {
	f := ListFilter{
		Search:      c.QueryParam("q"),
		CategoryKey: c.QueryParam("category"),
		Type:        c.QueryParam("type"),
		SortBy:      c.QueryParam("sort"),
		SortDir:     c.QueryParam("sort_dir"),
	}

	if v := c.QueryParam("in_stock"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid in_stock value"})
		}
		f.InStock = &b
	}

	if v := c.QueryParam("has_price"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid has_price value"})
		}
		f.HasPrice = &b
	}

	if v := c.QueryParam("min_price"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid min_price value"})
		}
		f.MinPrice = &n
	}

	if v := c.QueryParam("max_price"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid max_price value"})
		}
		f.MaxPrice = &n
	}

	f.Page, _ = strconv.Atoi(c.QueryParam("page"))
	f.Limit, _ = strconv.Atoi(c.QueryParam("limit"))

	result, err := h.svc.ListProducts(c.Request().Context(), f)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to load products"})
	}

	return c.JSON(http.StatusOK, result)
}

// @Summary     Get product by ID
// @Tags        products
// @Produce     json
// @Param       id path string true "Product GUID (Ref_Key)"
// @Success     200 {object} Product
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /products/{id} [get]
func (h *Handler) getProduct(c echo.Context) error {
	id := c.Param("id")

	p, err := h.svc.GetProduct(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "product not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to load product"})
	}

	return c.JSON(http.StatusOK, p)
}

// @Summary     List categories
// @Tags        products
// @Produce     json
// @Success     200 {array} Category
// @Failure     500 {object} map[string]string
// @Router      /categories [get]
func (h *Handler) listCategories(c echo.Context) error {
	cats, err := h.svc.ListCategories(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to load categories"})
	}

	return c.JSON(http.StatusOK, cats)
}
