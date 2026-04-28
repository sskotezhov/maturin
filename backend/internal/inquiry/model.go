package inquiry

import "time"

type Status string

const (
	StatusNew       Status = "new"
	StatusContacted Status = "contacted"
	StatusClosed    Status = "closed"
)

func ValidStatus(status string) bool {
	switch Status(status) {
	case StatusNew, StatusContacted, StatusClosed:
		return true
	default:
		return false
	}
}

type Inquiry struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	PhoneDigits     string    `json:"-"`
	Comment         string    `json:"comment"`
	Source          string    `json:"source"`
	PageURL         string    `json:"page_url"`
	ConsentAccepted bool      `json:"consent_accepted"`
	Status          Status    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
