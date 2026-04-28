package order

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	mw "github.com/sskotezhov/maturin/pkg/middleware"
	"github.com/sskotezhov/maturin/pkg/roles"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterCart(g *echo.Group) {
	g.GET("", h.getCart)
	g.POST("/items", h.addItem)
	g.PATCH("/items/:id", h.updateItem)
	g.DELETE("/items/:id", h.deleteItem)
	g.POST("/submit", h.submit)
}

func (h *Handler) RegisterOrders(g *echo.Group) {
	g.GET("", h.getOrders)
	g.GET("/:id", h.getOrder)
	g.DELETE("/:id", h.cancelOrder)
	g.POST("/:id/approve", h.approveOrder)
	g.GET("/:id/messages", h.getMessages)
	g.POST("/:id/messages", h.sendMessage)
}

type addItemRequest struct {
	ProductID     string   `json:"product_id"`
	ProductName   string   `json:"product_name"`
	ProductCode   string   `json:"product_code"`
	Quantity      int      `json:"quantity"`
	PriceSnapshot *float64 `json:"price_snapshot"`
	Comment       string   `json:"comment"`
}

type updateItemRequest struct {
	Quantity int    `json:"quantity"`
	Comment  string `json:"comment"`
}

type approveRequest struct {
	TotalPrice float64 `json:"total_price"`
}

type sendMessageRequest struct {
	Text string `json:"text"`
}

type orderItemResponse struct {
	ID            uint     `json:"id"`
	ProductID     string   `json:"product_id"`
	ProductName   string   `json:"product_name"`
	ProductCode   string   `json:"product_code"`
	Quantity      int      `json:"quantity"`
	PriceSnapshot *float64 `json:"price_snapshot"`
	Comment       string   `json:"comment"`
}

type orderResponse struct {
	ID             uint                `json:"id"`
	UserID         uint                `json:"user_id"`
	Status         string              `json:"status"`
	ResponseStatus string              `json:"response_status"`
	TotalPrice     *float64            `json:"total_price"`
	Items          []orderItemResponse `json:"items"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

type messageResponse struct {
	ID        uint   `json:"id"`
	OrderID   uint   `json:"order_id"`
	UserID    uint   `json:"user_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func toOrderResponse(o *Order) orderResponse {
	items := make([]orderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = orderItemResponse{
			ID:            item.ID,
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,
			ProductCode:   item.ProductCode,
			Quantity:      item.Quantity,
			PriceSnapshot: item.PriceSnapshot,
			Comment:       item.Comment,
		}
	}
	return orderResponse{
		ID:             o.ID,
		UserID:         o.UserID,
		Status:         string(o.Status),
		ResponseStatus: responseStatus(o),
		TotalPrice:     o.TotalPrice,
		Items:          items,
		CreatedAt:      o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func responseStatus(o *Order) string {
	if o.ResponseStatus == "" {
		return string(ResponseNone)
	}
	return string(o.ResponseStatus)
}

func toMessageResponse(m *Message) messageResponse {
	return messageResponse{
		ID:        m.ID,
		OrderID:   m.OrderID,
		UserID:    m.UserID,
		Text:      m.Text,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Handlers

// @Summary     Get current draft cart
// @Tags        cart
// @Produce     json
// @Success     200 {object} orderResponse
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /cart [get]
func (h *Handler) getCart(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)

	order, err := h.svc.GetCart(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "no active cart"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, toOrderResponse(order))
}

// @Summary     Add item to cart
// @Tags        cart
// @Accept      json
// @Produce     json
// @Param       body body addItemRequest true "Item details"
// @Success     200 {object} orderResponse
// @Failure     400 {object} map[string]string
// @Security    BearerAuth
// @Router      /cart/items [post]
func (h *Handler) addItem(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)

	var req addItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if req.ProductID == "" || req.ProductName == "" || req.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "product_id, product_name and quantity are required"})
	}

	order, err := h.svc.AddItem(c.Request().Context(), userID, AddItemInput{
		ProductID:     req.ProductID,
		ProductName:   req.ProductName,
		ProductCode:   req.ProductCode,
		Quantity:      req.Quantity,
		PriceSnapshot: req.PriceSnapshot,
		Comment:       req.Comment,
	})
	if err != nil {
		if errors.Is(err, ErrCommentRequired) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, toOrderResponse(order))
}

// @Summary     Update cart item
// @Tags        cart
// @Accept      json
// @Produce     json
// @Param       id path int true "Item ID"
// @Param       body body updateItemRequest true "Quantity and comment"
// @Success     200 {object} orderResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /cart/items/{id} [patch]
func (h *Handler) updateItem(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	itemID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid item id"})
	}

	var req updateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if req.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "quantity must be positive"})
	}

	order, err := h.svc.UpdateItem(c.Request().Context(), userID, itemID, UpdateItemInput{
		Quantity: req.Quantity,
		Comment:  req.Comment,
	})
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, toOrderResponse(order))
}

// @Summary     Delete cart item
// @Tags        cart
// @Param       id path int true "Item ID"
// @Success     204
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /cart/items/{id} [delete]
func (h *Handler) deleteItem(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	itemID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid item id"})
	}

	if err := h.svc.DeleteItem(c.Request().Context(), userID, itemID); err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary     Submit cart (draft → submitted)
// @Tags        cart
// @Produce     json
// @Success     200 {object} orderResponse
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /cart/submit [post]
func (h *Handler) submit(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)

	order, err := h.svc.Submit(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, toOrderResponse(order))
}

// @Summary     List orders
// @Tags        orders
// @Produce     json
// @Param       status   query string false "Filter by status (draft, submitted, approved, cancelled); manager/admin see draft only for themselves"
// @Param       user_id  query int    false "Filter by client user ID (manager/admin only)"
// @Param       page     query int    false "Page number (default 1)"
// @Param       limit    query int    false "Items per page (default 20)"
// @Success     200 {array} orderResponse
// @Security    BearerAuth
// @Router      /orders [get]
func (h *Handler) getOrders(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	role := c.Get(mw.ContextRole).(string)

	f := Filter{
		Status: c.QueryParam("status"),
		Page:   queryInt(c, "page", 1),
		Limit:  queryInt(c, "limit", 20),
	}
	if role == string(roles.RoleManager) || role == string(roles.RoleAdmin) {
		if uid := queryInt(c, "user_id", 0); uid > 0 {
			f.UserID = uint(uid)
		}
	}

	orders, err := h.svc.GetOrders(c.Request().Context(), userID, role, f)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	resp := make([]orderResponse, len(orders))
	for i, o := range orders {
		resp[i] = toOrderResponse(o)
	}
	return c.JSON(http.StatusOK, resp)
}

// @Summary     Get order details
// @Tags        orders
// @Produce     json
// @Param       id path int true "Order ID"
// @Success     200 {object} orderResponse
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /orders/{id} [get]
func (h *Handler) getOrder(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	role := c.Get(mw.ContextRole).(string)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}

	order, err := h.svc.GetOrder(c.Request().Context(), userID, role, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, toOrderResponse(order))
}

// @Summary     Cancel order (client only, before approved)
// @Tags        orders
// @Param       id path int true "Order ID"
// @Success     204
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /orders/{id} [delete]
func (h *Handler) cancelOrder(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	role := c.Get(mw.ContextRole).(string)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}

	if err := h.svc.CancelOrder(c.Request().Context(), userID, role, orderID); err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary     Approve order and set final price (manager/admin)
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       id path int true "Order ID"
// @Param       body body approveRequest true "Final price"
// @Success     200 {object} orderResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /orders/{id}/approve [post]
func (h *Handler) approveOrder(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	role := c.Get(mw.ContextRole).(string)

	if role != string(roles.RoleManager) && role != string(roles.RoleAdmin) {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden"})
	}

	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}

	var req approveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if req.TotalPrice <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "total_price must be positive"})
	}

	order, err := h.svc.ApproveOrder(c.Request().Context(), userID, role, orderID, req.TotalPrice)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, toOrderResponse(order))
}

// @Summary     Get order chat messages
// @Tags        orders
// @Produce     json
// @Param       id path int true "Order ID"
// @Success     200 {array} messageResponse
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /orders/{id}/messages [get]
func (h *Handler) getMessages(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	role := c.Get(mw.ContextRole).(string)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}

	msgs, err := h.svc.GetMessages(c.Request().Context(), userID, role, orderID)
	if err != nil {
		return mapServiceError(c, err)
	}

	resp := make([]messageResponse, len(msgs))
	for i := range msgs {
		resp[i] = toMessageResponse(&msgs[i])
	}
	return c.JSON(http.StatusOK, resp)
}

// @Summary     Send message to order chat
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       id path int true "Order ID"
// @Param       body body sendMessageRequest true "Message text"
// @Success     201 {object} messageResponse
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Security    BearerAuth
// @Router      /orders/{id}/messages [post]
func (h *Handler) sendMessage(c echo.Context) error {
	userID := c.Get(mw.ContextUserID).(uint)
	role := c.Get(mw.ContextRole).(string)
	orderID, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid order id"})
	}

	var req sendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if req.Text == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "text is required"})
	}

	msg, err := h.svc.SendMessage(c.Request().Context(), userID, role, orderID, req.Text)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, toMessageResponse(msg))
}

func parseID(c echo.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func queryInt(c echo.Context, param string, defaultVal int) int {
	v, err := strconv.Atoi(c.QueryParam(param))
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	case errors.Is(err, ErrForbidden):
		return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
	case errors.Is(err, ErrInvalidStatus):
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	case errors.Is(err, ErrCommentRequired):
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}
}
