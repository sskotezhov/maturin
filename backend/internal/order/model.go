package order

import "time"

type Status string
type ResponseStatus string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusApproved  Status = "approved"
	StatusCancelled Status = "cancelled"
)

const (
	ResponseNone           ResponseStatus = "none"
	ResponseWaitingClient  ResponseStatus = "waiting_client"
	ResponseWaitingManager ResponseStatus = "waiting_manager"
)

type Order struct {
	ID             uint
	UserID         uint
	Status         Status
	ResponseStatus ResponseStatus
	TotalPrice     *float64
	Items          []Item
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Item struct {
	ID            uint
	OrderID       uint
	ProductID     string
	ProductName   string
	ProductCode   string
	Quantity      int
	PriceSnapshot *float64
	Comment       string
	CreatedAt     time.Time
}

type Message struct {
	ID        uint
	OrderID   uint
	UserID    uint
	Text      string
	CreatedAt time.Time
}
