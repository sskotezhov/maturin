package staff

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/sskotezhov/maturin/internal/inquiry"
	"github.com/sskotezhov/maturin/internal/order"
	"github.com/sskotezhov/maturin/internal/user"
	mw "github.com/sskotezhov/maturin/pkg/middleware"
	"github.com/sskotezhov/maturin/pkg/roles"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/clients", h.listClients)
	g.GET("/clients/:id", h.getClient)
	g.GET("/inquiries", h.listInquiries)
	g.GET("/inquiries/:id", h.getInquiry)
	g.PATCH("/inquiries/:id/status", h.changeInquiryStatus)
	g.GET("/dashboard", h.getDashboard)
	g.POST("/orders/:id/items", h.staffAddItem)
	g.PATCH("/orders/:id/items/:itemID", h.staffUpdateItem)
	g.DELETE("/orders/:id/items/:itemID", h.staffDeleteItem)

	adminOnly := g.Group("", mw.RequireRoles(roles.RoleAdmin))
	adminOnly.PATCH("/clients/:id/role", h.changeRole)
	adminOnly.POST("/cache/refresh", h.refreshCache)
}

type clientView struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Role          string `json:"role"`
	LastName      string `json:"last_name"`
	FirstName     string `json:"first_name"`
	MiddleName    string `json:"middle_name"`
	Phone         string `json:"phone"`
	Telegram      string `json:"telegram"`
	CompanyName   string `json:"company_name"`
	INN           string `json:"inn"`
	Comment       string `json:"comment"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type clientListResponse struct {
	Items []clientView `json:"items"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type orderItemView struct {
	ID            uint     `json:"id"`
	ProductID     string   `json:"product_id"`
	ProductName   string   `json:"product_name"`
	ProductCode   string   `json:"product_code"`
	Quantity      int      `json:"quantity"`
	PriceSnapshot *float64 `json:"price_snapshot"`
	Comment       string   `json:"comment"`
}

type userInfoView struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	LastName    string `json:"last_name"`
	FirstName   string `json:"first_name"`
	Phone       string `json:"phone"`
	CompanyName string `json:"company_name"`
}

type orderView struct {
	ID             uint            `json:"id"`
	User           *userInfoView   `json:"user"`
	Status         string          `json:"status"`
	ResponseStatus string          `json:"response_status"`
	TotalPrice     *float64        `json:"total_price"`
	OnecRef        *string         `json:"onec_ref"`
	OnecNumber     *string         `json:"onec_number"`
	Items          []orderItemView `json:"items"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

type clientDetailsResponse struct {
	Client      clientView  `json:"client"`
	Orders      []orderView `json:"orders"`
	OrdersCount int         `json:"orders_count"`
}

type dashboardResponse struct {
	OrdersByStatus      map[string]int `json:"orders_by_status"`
	StaleSubmittedCount int            `json:"stale_submitted_count"`
	CatalogSyncedAt     *string        `json:"catalog_synced_at"`
}

type inquiryView struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Phone           string `json:"phone"`
	Comment         string `json:"comment"`
	Source          string `json:"source"`
	PageURL         string `json:"page_url"`
	ConsentAccepted bool   `json:"consent_accepted"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type inquiryListResponse struct {
	Items []inquiryView `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

type changeRoleRequest struct {
	Role string `json:"role"`
}

type changeInquiryStatusRequest struct {
	Status string `json:"status"`
}

type cacheRefreshResponse struct {
	RefreshedAt     string `json:"refreshed_at"`
	ProductsCount   int    `json:"products_count"`
	CategoriesCount int    `json:"categories_count"`
}

func toClientView(u *user.User) clientView {
	return clientView{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Role:          string(u.Role),
		LastName:      u.LastName,
		FirstName:     u.FirstName,
		MiddleName:    u.MiddleName,
		Phone:         u.Phone,
		Telegram:      u.Telegram,
		CompanyName:   u.CompanyName,
		INN:           u.INN,
		Comment:       u.Comment,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toOrderView(o *order.Order) orderView {
	items := make([]orderItemView, len(o.Items))
	for i, item := range o.Items {
		items[i] = orderItemView{
			ID:            item.ID,
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,
			ProductCode:   item.ProductCode,
			Quantity:      item.Quantity,
			PriceSnapshot: item.PriceSnapshot,
			Comment:       item.Comment,
		}
	}
	var u *userInfoView
	if o.User != nil {
		u = &userInfoView{
			ID:          o.User.ID,
			Email:       o.User.Email,
			LastName:    o.User.LastName,
			FirstName:   o.User.FirstName,
			Phone:       o.User.Phone,
			CompanyName: o.User.CompanyName,
		}
	}
	return orderView{
		ID:             o.ID,
		User:           u,
		Status:         string(o.Status),
		ResponseStatus: orderResponseStatus(o),
		TotalPrice:     o.TotalPrice,
		OnecRef:        o.OnecRef,
		OnecNumber:     o.OnecNumber,
		Items:          items,
		CreatedAt:      o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func orderResponseStatus(o *order.Order) string {
	if o.ResponseStatus == "" {
		return string(order.ResponseNone)
	}
	return string(o.ResponseStatus)
}

func toInquiryView(i *inquiry.Inquiry) inquiryView {
	return inquiryView{
		ID:              i.ID,
		Name:            i.Name,
		Phone:           i.Phone,
		Comment:         i.Comment,
		Source:          i.Source,
		PageURL:         i.PageURL,
		ConsentAccepted: i.ConsentAccepted,
		Status:          string(i.Status),
		CreatedAt:       i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// @Summary     List/search clients (manager/admin)
// @Tags        staff
// @Produce     json
// @Param       q              query string false "Search across email/name/company/phone/inn/telegram"
// @Param       role           query string false "Filter by role (client/manager/admin)"
// @Param       email_verified query bool   false "Filter by email_verified"
// @Param       page           query int    false "Page (default 1)"
// @Param       limit          query int    false "Limit (default 20)"
// @Success     200 {object} clientListResponse
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/clients [get]
func (h *Handler) listClients(c echo.Context) error {
	f := ClientFilter{
		Q:     c.QueryParam("q"),
		Role:  c.QueryParam("role"),
		Page:  queryInt(c, "page", 1),
		Limit: queryInt(c, "limit", 20),
	}
	if v := c.QueryParam("email_verified"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "email_verified must be true/false"})
		}
		f.EmailVerified = &b
	}
	if f.Role != "" && !roles.Valid(f.Role) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid role"})
	}

	users, total, err := h.svc.ListClients(c.Request().Context(), f)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	items := make([]clientView, len(users))
	for i, u := range users {
		items[i] = toClientView(u)
	}

	return c.JSON(http.StatusOK, clientListResponse{
		Items: items,
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	})
}

// @Summary     List/search public inquiries (manager/admin)
// @Tags        staff
// @Produce     json
// @Param       q      query string false "Search by name, phone or comment"
// @Param       status query string false "Filter by status: new | contacted | closed"
// @Param       page   query int    false "Page (default 1)"
// @Param       limit  query int    false "Limit (default 20)"
// @Success     200 {object} inquiryListResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/inquiries [get]
func (h *Handler) listInquiries(c echo.Context) error {
	f := inquiry.Filter{
		Q:      c.QueryParam("q"),
		Status: c.QueryParam("status"),
		Page:   queryInt(c, "page", 1),
		Limit:  queryInt(c, "limit", 20),
	}

	items, total, err := h.svc.ListInquiries(c.Request().Context(), f)
	if err != nil {
		if errors.Is(err, ErrInvalidInquiryStatus) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid status"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	resp := make([]inquiryView, len(items))
	for i, item := range items {
		resp[i] = toInquiryView(item)
	}

	return c.JSON(http.StatusOK, inquiryListResponse{
		Items: resp,
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	})
}

// @Summary     Get public inquiry details (manager/admin)
// @Tags        staff
// @Produce     json
// @Param       id path int true "Inquiry ID"
// @Success     200 {object} inquiryView
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/inquiries/{id} [get]
func (h *Handler) getInquiry(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	item, err := h.svc.GetInquiry(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, toInquiryView(item))
}

// @Summary     Change public inquiry status (manager/admin)
// @Tags        staff
// @Accept      json
// @Produce     json
// @Param       id   path int                   true "Inquiry ID"
// @Param       body body changeInquiryStatusRequest true "New status"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/inquiries/{id}/status [patch]
func (h *Handler) changeInquiryStatus(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	var req changeInquiryStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	if err := h.svc.ChangeInquiryStatus(c.Request().Context(), id, inquiry.Status(req.Status)); err != nil {
		switch {
		case errors.Is(err, ErrInvalidInquiryStatus):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid status"})
		case errors.Is(err, ErrNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary     Get client details with recent orders (manager/admin)
// @Tags        staff
// @Produce     json
// @Param       id path int true "User ID"
// @Success     200 {object} clientDetailsResponse
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/clients/{id} [get]
func (h *Handler) getClient(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	details, err := h.svc.GetClient(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	orders := make([]orderView, len(details.Orders))
	for i, o := range details.Orders {
		orders[i] = toOrderView(o)
	}

	return c.JSON(http.StatusOK, clientDetailsResponse{
		Client:      toClientView(details.Client),
		Orders:      orders,
		OrdersCount: details.OrdersCount,
	})
}

// @Summary     Staff dashboard (manager/admin)
// @Tags        staff
// @Produce     json
// @Success     200 {object} dashboardResponse
// @Security    BearerAuth
// @Router      /staff/dashboard [get]
func (h *Handler) getDashboard(c echo.Context) error {
	d, err := h.svc.Dashboard(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}
	var syncedAt *string
	if d.CatalogSyncedAt != nil {
		s := d.CatalogSyncedAt.Format("2006-01-02T15:04:05Z07:00")
		syncedAt = &s
	}
	return c.JSON(http.StatusOK, dashboardResponse{
		OrdersByStatus:      d.OrdersByStatus,
		StaleSubmittedCount: d.StaleSubmittedCount,
		CatalogSyncedAt:     syncedAt,
	})
}

// @Summary     Change user role (admin only)
// @Tags        staff
// @Accept      json
// @Produce     json
// @Param       id   path int                true "User ID"
// @Param       body body changeRoleRequest  true "New role"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     422 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/clients/{id}/role [patch]
func (h *Handler) changeRole(c echo.Context) error {
	actorID := c.Get(mw.ContextUserID).(uint)
	targetID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	var req changeRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	if err := h.svc.ChangeRole(c.Request().Context(), actorID, targetID, roles.Role(req.Role)); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		case errors.Is(err, ErrInvalidRole):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid role"})
		case errors.Is(err, ErrSelfRoleChange):
			return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": "cannot change own role"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary     Refresh 1C catalog cache (admin only)
// @Tags        staff
// @Produce     json
// @Success     200 {object} cacheRefreshResponse
// @Failure     500 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/cache/refresh [post]
func (h *Handler) refreshCache(c echo.Context) error {
	stats, err := h.svc.RefreshCatalog(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "refresh failed"})
	}
	return c.JSON(http.StatusOK, cacheRefreshResponse{
		RefreshedAt:     stats.RefreshedAt.Format("2006-01-02T15:04:05Z07:00"),
		ProductsCount:   stats.ProductsCount,
		CategoriesCount: stats.CategoriesCount,
	})
}

type staffAddItemRequest struct {
	ProductID     string   `json:"product_id"`
	ProductName   string   `json:"product_name"`
	ProductCode   string   `json:"product_code"`
	Quantity      int      `json:"quantity"`
	PriceSnapshot *float64 `json:"price_snapshot"`
	Comment       string   `json:"comment"`
}

type staffUpdateItemRequest struct {
	Quantity int    `json:"quantity"`
	Comment  string `json:"comment"`
}

// @Summary     Add item to client's order (manager/admin)
// @Tags        staff
// @Accept      json
// @Produce     json
// @Param       id   path int                  true "Order ID"
// @Param       body body staffAddItemRequest  true "Item to add"
// @Success     200 {object} orderView
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/orders/{id}/items [post]
func (h *Handler) staffAddItem(c echo.Context) error {
	actorID := c.Get(mw.ContextUserID).(uint)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}

	var req staffAddItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if req.ProductID == "" || req.ProductName == "" || req.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "product_id, product_name and quantity are required"})
	}

	ord, err := h.svc.StaffAddItem(c.Request().Context(), actorID, orderID, order.AddItemInput{
		ProductID:     req.ProductID,
		ProductName:   req.ProductName,
		ProductCode:   req.ProductCode,
		Quantity:      req.Quantity,
		PriceSnapshot: req.PriceSnapshot,
		Comment:       req.Comment,
	})
	if err != nil {
		switch {
		case errors.Is(err, order.ErrNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "order not found"})
		case errors.Is(err, order.ErrOrderNotEditable):
			return c.JSON(http.StatusConflict, echo.Map{"error": "order cannot be edited in current status"})
		case errors.Is(err, order.ErrCommentRequired):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "comment required for items without price"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusOK, toOrderView(ord))
}

// @Summary     Update item in client's order (manager/admin)
// @Tags        staff
// @Accept      json
// @Produce     json
// @Param       id     path int                    true "Order ID"
// @Param       itemID path int                    true "Item ID"
// @Param       body   body staffUpdateItemRequest true "Updated item"
// @Success     200 {object} orderView
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/orders/{id}/items/{itemID} [patch]
func (h *Handler) staffUpdateItem(c echo.Context) error {
	actorID := c.Get(mw.ContextUserID).(uint)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}
	itemID, err := parseID(c, "itemID")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid item id"})
	}

	var req staffUpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if req.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "quantity must be positive"})
	}

	ord, err := h.svc.StaffUpdateItem(c.Request().Context(), actorID, orderID, itemID, order.UpdateItemInput{
		Quantity: req.Quantity,
		Comment:  req.Comment,
	})
	if err != nil {
		switch {
		case errors.Is(err, order.ErrNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
		case errors.Is(err, order.ErrOrderNotEditable):
			return c.JSON(http.StatusConflict, echo.Map{"error": "order cannot be edited in current status"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusOK, toOrderView(ord))
}

// @Summary     Delete item from client's order (manager/admin)
// @Tags        staff
// @Param       id     path int true "Order ID"
// @Param       itemID path int true "Item ID"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Security    BearerAuth
// @Router      /staff/orders/{id}/items/{itemID} [delete]
func (h *Handler) staffDeleteItem(c echo.Context) error {
	actorID := c.Get(mw.ContextUserID).(uint)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}
	itemID, err := parseID(c, "itemID")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid item id"})
	}

	if err := h.svc.StaffDeleteItem(c.Request().Context(), actorID, orderID, itemID); err != nil {
		switch {
		case errors.Is(err, order.ErrNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
		case errors.Is(err, order.ErrOrderNotEditable):
			return c.JSON(http.StatusConflict, echo.Map{"error": "order cannot be edited in current status"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func parseID(c echo.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func queryInt(c echo.Context, param string, def int) int {
	v, err := strconv.Atoi(c.QueryParam(param))
	if err != nil || v < 0 {
		return def
	}
	return v
}
