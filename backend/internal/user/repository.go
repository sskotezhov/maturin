package user

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/sskotezhov/maturin/pkg/roles"
)

type Filter struct {
	Q             string
	Role          string
	EmailVerified *bool
	Page          int
	Limit         int
}

// Repository defines the interface for user data access.
type Repository interface {
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAllByRole(ctx context.Context, role roles.Role) ([]*User, error)
	FindFiltered(ctx context.Context, f Filter) ([]*User, int, error)
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

func (r *repository) FindAllByRole(ctx context.Context, role roles.Role) ([]*User, error) {
	var recs []userRecord
	if err := r.db.WithContext(ctx).Where("role = ?", string(role)).Find(&recs).Error; err != nil {
		return nil, err
	}
	users := make([]*User, len(recs))
	for i, rec := range recs {
		users[i] = toEntity(rec)
	}
	return users, nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	rec := toDB(u)
	return r.db.WithContext(ctx).Save(&rec).Error
}

func (r *repository) FindFiltered(ctx context.Context, f Filter) ([]*User, int, error) {
	base := r.db.WithContext(ctx).Model(&userRecord{})

	if f.Q != "" {
		like := "%" + strings.ToLower(f.Q) + "%"
		digits := digitsOnly(f.Q)
		if digits != "" {
			base = base.Where(`
				LOWER(email) LIKE ? OR
				LOWER(last_name || ' ' || first_name) LIKE ? OR
				LOWER(company_name) LIKE ? OR
				LOWER(telegram) LIKE ? OR
				inn LIKE ? OR
				regexp_replace(phone, '\D', '', 'g') LIKE ?
			`, like, like, like, like, like, "%"+digits+"%")
		} else {
			base = base.Where(`
				LOWER(email) LIKE ? OR
				LOWER(last_name || ' ' || first_name) LIKE ? OR
				LOWER(company_name) LIKE ? OR
				LOWER(telegram) LIKE ? OR
				inn LIKE ?
			`, like, like, like, like, like)
		}
	}
	if f.Role != "" {
		base = base.Where("role = ?", f.Role)
	}
	if f.EmailVerified != nil {
		base = base.Where("email_verified = ?", *f.EmailVerified)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := base.Order("created_at desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset((f.Page - 1) * f.Limit)
	}

	var recs []userRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*User, len(recs))
	for i, rec := range recs {
		out[i] = toEntity(rec)
	}
	return out, int(total), nil
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
