package user

import "time"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleClient  Role = "client"
)

type User struct {
	ID            uint
	Email         string
	EmailVerified bool
	PasswordHash  string
	Role          Role

	// Profile
	LastName    string
	FirstName   string
	MiddleName  string
	Phone       string
	Telegram    string
	CompanyName string
	INN         string
	Comment     string

	CreatedAt time.Time
	UpdatedAt time.Time
}
