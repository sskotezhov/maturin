package user

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository defines the interface for user data access.
type Repository interface {
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}

type userRecord struct {
	ID            uint   `gorm:"primaryKey"`
	Email         string `gorm:"uniqueIndex;not null"`
	EmailVerified bool   `gorm:"not null;default:false"`
	PasswordHash  string `gorm:"not null"`
	Role          string `gorm:"type:varchar(20);not null;default:'client'"`

	LastName    string `gorm:"not null;default:''"`
	FirstName   string `gorm:"not null;default:''"`
	MiddleName  string `gorm:"not null;default:''"`
	Phone       string `gorm:"not null;default:''"`
	Telegram    string `gorm:"not null;default:''"`
	CompanyName string `gorm:"not null;default:''"`
	INN         string `gorm:"not null;default:''"`
	Comment     string `gorm:"not null;default:''"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (userRecord) TableName() string { return "users" }

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&userRecord{})
}

func toEntity(r userRecord) *User {
	return &User{
		ID:            r.ID,
		Email:         r.Email,
		EmailVerified: r.EmailVerified,
		PasswordHash:  r.PasswordHash,
		Role:          Role(r.Role),
		LastName:      r.LastName,
		FirstName:     r.FirstName,
		MiddleName:    r.MiddleName,
		Phone:         r.Phone,
		Telegram:      r.Telegram,
		CompanyName:   r.CompanyName,
		INN:           r.INN,
		Comment:       r.Comment,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func toDB(u *User) userRecord {
	return userRecord{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		PasswordHash:  u.PasswordHash,
		Role:          string(u.Role),
		LastName:      u.LastName,
		FirstName:     u.FirstName,
		MiddleName:    u.MiddleName,
		Phone:         u.Phone,
		Telegram:      u.Telegram,
		CompanyName:   u.CompanyName,
		INN:           u.INN,
		Comment:       u.Comment,
	}
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var rec userRecord
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		return nil, err
	}
	return toEntity(rec), nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var rec userRecord
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&rec).Error; err != nil {
		return nil, err
	}
	return toEntity(rec), nil
}

func (r *repository) Create(ctx context.Context, u *User) error {
	rec := toDB(u)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	u.ID = rec.ID
	u.CreatedAt = rec.CreatedAt
	u.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	rec := toDB(u)
	return r.db.WithContext(ctx).Save(&rec).Error
}
