package slot

import "time"

type CallSlot struct {
	ID        uint      `json:"id"`
	ManagerID uint      `json:"manager_id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	IsBooked  bool      `json:"is_booked"`
	CreatedAt time.Time `json:"created_at"`
}

type CallBooking struct {
	ID          uint      `json:"id"`
	SlotID      uint      `json:"slot_id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	PhoneDigits string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}
