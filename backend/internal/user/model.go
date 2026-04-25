package user

import (
	"time"

	"github.com/sskotezhov/maturin/pkg/roles"
)

type Role = roles.Role

const (
	RoleAdmin   = roles.RoleAdmin
	RoleManager = roles.RoleManager
	RoleClient  = roles.RoleClient
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
