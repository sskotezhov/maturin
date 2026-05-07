package slot

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	mw "github.com/sskotezhov/maturin/pkg/middleware"
	"github.com/sskotezhov/maturin/pkg/roles"
	"gorm.io/gorm"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublic(g *echo.Group) {
	g.GET("", h.listAvailable)
	g.POST("/:id/book", h.book)
}

func (h *Handler) RegisterStaff(g *echo.Group) {
	g.GET("", h.staffListSlots)
	g.GET("/bookings", h.staffListBookings)

	adminOnly := g.Group("", mw.RequireRoles(roles.RoleAdmin))
	adminOnly.POST("/ranges", h.createRanges)
	adminOnly.DELETE("/:id", h.deleteSlot)
}

// — Public types —

type slotView struct {
	ID        uint      `json:"id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	ManagerID uint      `json:"manager_id"`
}

type bookRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Comment string `json:"comment"`
}

type bookResponse struct {
	ID uint `json:"id"`
}

// — Staff types —

type rangeInput struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	ManagerID uint      `json:"manager_id"`
}

type createRangesRequest struct {
	Ranges []rangeInput `json:"ranges"`
}

type createRangesResponse struct {
	Created int `json:"created"`
}

type staffSlotView struct {
	ID        uint      `json:"id"`
	ManagerID uint      `json:"manager_id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	IsBooked  bool      `json:"is_booked"`
	CreatedAt time.Time `json:"created_at"`
}

type staffSlotsResponse struct {
	Items []*staffSlotView `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type bookingView struct {
	ID        uint      `json:"id"`
	SlotID    uint      `json:"slot_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type staffBookingsResponse struct {
	Items []*bookingView `json:"items"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// @Summary     List available call slots
// @Tags        slots
// @Produce     json
// @Param       date       query string false "Filter by date (YYYY-MM-DD)"
// @Param       manager_id query int    false "Filter by manager ID"
// @Success     200 {array} slotView
// @Failure     500 {object} map[string]string
// @Router      /slots [get]
func (h *Handler) listAvailable(c echo.Context) error {
	f := AvailableFilter{}
	if d := c.QueryParam("date"); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err == nil {
			f.Date = &t
		}
	}
	if mid := c.QueryParam("manager_id"); mid != "" {
		id, err := strconv.ParseUint(mid, 10, 64)
		if err == nil {
			uid := uint(id)
			f.ManagerID = &uid
		}
	}

	slots, err := h.svc.ListAvailable(c.Request().Context(), f)
	if err != nil {
		slog.Error("slot: list available failed", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	views := make([]*slotView, len(slots))
	for i, s := range slots {
		views[i] = &slotView{ID: s.ID, StartAt: s.StartAt, EndAt: s.EndAt, ManagerID: s.ManagerID}
	}
	return c.JSON(http.StatusOK, views)
}

// @Summary     Book a call slot
// @Tags        slots
// @Accept      json
// @Produce     json
// @Param       id   path int         true "Slot ID"
// @Param       body body bookRequest true "Booking info"
// @Success     201 {object} bookResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /slots/{id}/book [post]
func (h *Handler) book(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid slot id"})
	}

	var req bookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	booking, err := h.svc.Book(c.Request().Context(), id, BookInput{Name: req.Name, Phone: req.Phone, Comment: req.Comment})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidName):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid name"})
		case errors.Is(err, ErrInvalidPhone):
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid phone"})
		case errors.Is(err, ErrSlotNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "slot not found"})
		case errors.Is(err, ErrSlotAlreadyBooked):
			return c.JSON(http.StatusConflict, echo.Map{"error": "slot already booked"})
		default:
			slog.Error("slot: book failed", "slot_id", id, "err", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.JSON(http.StatusCreated, bookResponse{ID: booking.ID})
}

// @Summary     Create call slot ranges (admin only)
// @Tags        slots
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createRangesRequest true "Time ranges"
// @Success     201 {object} createRangesResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /staff/slots/ranges [post]
func (h *Handler) createRanges(c echo.Context) error {
	var req createRangesRequest
	if err := c.Bind(&req); err != nil || len(req.Ranges) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	inputs := make([]RangeInput, len(req.Ranges))
	for i, r := range req.Ranges {
		inputs[i] = RangeInput{Start: r.Start, End: r.End, ManagerID: r.ManagerID}
	}

	count, err := h.svc.CreateRanges(c.Request().Context(), inputs)
	if err != nil {
		if errors.Is(err, ErrInvalidRange) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "range too small, no 30-min slots fit"})
		}
		slog.Error("slot: create ranges failed", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	return c.JSON(http.StatusCreated, createRangesResponse{Created: count})
}

// @Summary     List slots (staff)
// @Tags        slots
// @Produce     json
// @Security    BearerAuth
// @Param       status     query string false "free | booked | (empty = all)"
// @Param       date       query string false "Filter by date (YYYY-MM-DD)"
// @Param       manager_id query int    false "Filter by manager ID"
// @Param       page       query int    false "Page (default 1)"
// @Param       limit      query int    false "Limit (default 20)"
// @Success     200 {object} staffSlotsResponse
// @Failure     500 {object} map[string]string
// @Router      /staff/slots [get]
func (h *Handler) staffListSlots(c echo.Context) error {
	f := Filter{
		Status: c.QueryParam("status"),
		Page:   queryInt(c, "page", 1),
		Limit:  queryInt(c, "limit", 20),
	}
	if d := c.QueryParam("date"); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err == nil {
			f.Date = &t
		}
	}
	if mid := c.QueryParam("manager_id"); mid != "" {
		id, err := strconv.ParseUint(mid, 10, 64)
		if err == nil {
			uid := uint(id)
			f.ManagerID = &uid
		}
	}

	slots, total, err := h.svc.StaffListSlots(c.Request().Context(), f)
	if err != nil {
		slog.Error("slot: staff list failed", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	views := make([]*staffSlotView, len(slots))
	for i, s := range slots {
		views[i] = toStaffSlotView(s)
	}
	return c.JSON(http.StatusOK, staffSlotsResponse{Items: views, Total: total, Page: f.Page, Limit: f.Limit})
}

// @Summary     Delete a slot (admin only)
// @Tags        slots
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Slot ID"
// @Success     204
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /staff/slots/{id} [delete]
func (h *Handler) deleteSlot(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid slot id"})
	}

	if err := h.svc.StaffDeleteSlot(c.Request().Context(), id); err != nil {
		switch {
		case errors.Is(err, ErrSlotNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": "slot not found"})
		case errors.Is(err, ErrSlotBooked):
			return c.JSON(http.StatusConflict, echo.Map{"error": "slot is already booked"})
		default:
			slog.Error("slot: delete failed", "slot_id", id, "err", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary     List bookings (staff)
// @Tags        slots
// @Produce     json
// @Security    BearerAuth
// @Param       page  query int false "Page (default 1)"
// @Param       limit query int false "Limit (default 20)"
// @Success     200 {object} staffBookingsResponse
// @Failure     500 {object} map[string]string
// @Router      /staff/slots/bookings [get]
func (h *Handler) staffListBookings(c echo.Context) error {
	f := BookingFilter{
		Page:  queryInt(c, "page", 1),
		Limit: queryInt(c, "limit", 20),
	}

	bookings, total, err := h.svc.StaffListBookings(c.Request().Context(), f)
	if err != nil {
		slog.Error("slot: staff list bookings failed", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	views := make([]*bookingView, len(bookings))
	for i, b := range bookings {
		views[i] = &bookingView{
			ID:        b.ID,
			SlotID:    b.SlotID,
			Name:      b.Name,
			Phone:     b.Phone,
			CreatedAt: b.CreatedAt,
		}
	}
	return c.JSON(http.StatusOK, staffBookingsResponse{Items: views, Total: total, Page: f.Page, Limit: f.Limit})
}

func toStaffSlotView(s *CallSlot) *staffSlotView {
	return &staffSlotView{
		ID:        s.ID,
		ManagerID: s.ManagerID,
		StartAt:   s.StartAt,
		EndAt:     s.EndAt,
		IsBooked:  s.IsBooked,
		CreatedAt: s.CreatedAt,
	}
}

func parseID(c echo.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	return uint(id), err
}

func queryInt(c echo.Context, param string, def int) int {
	v, err := strconv.Atoi(c.QueryParam(param))
	if err != nil || v < 1 {
		return def
	}
	return v
}
