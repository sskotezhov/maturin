package inquiry

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Filter struct {
	Q      string
	Status string
	Page   int
	Limit  int
}

type Repository interface {
	Create(ctx context.Context, item *Inquiry) error
	FindByID(ctx context.Context, id uint) (*Inquiry, error)
	FindFiltered(ctx context.Context, f Filter) ([]*Inquiry, int, error)
	UpdateStatus(ctx context.Context, id uint, status Status) error
}

type inquiryRecord struct {
	ID              uint   `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	Phone           string `gorm:"not null"`
	PhoneDigits     string `gorm:"not null;index"`
	Comment         string `gorm:"not null;default:''"`
	Source          string `gorm:"not null;default:''"`
	PageURL         string `gorm:"not null;default:''"`
	ConsentAccepted bool   `gorm:"not null;default:false"`
	Status          string `gorm:"type:varchar(20);not null;default:'new';index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (inquiryRecord) TableName() string { return "inquiries" }

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&inquiryRecord{})
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *Inquiry) error {
	rec := toDB(item)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*item = *toEntity(rec)
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*Inquiry, error) {
	var rec inquiryRecord
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		return nil, err
	}
	return toEntity(rec), nil
}

func (r *repository) FindFiltered(ctx context.Context, f Filter) ([]*Inquiry, int, error) {
	base := r.db.WithContext(ctx).Model(&inquiryRecord{})

	if f.Status != "" {
		base = base.Where("status = ?", f.Status)
	}
	if f.Q != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(f.Q)) + "%"
		digits := digitsOnly(f.Q)
		if digits != "" {
			base = base.Where(`
				LOWER(name) LIKE ? OR
				LOWER(comment) LIKE ? OR
				phone_digits LIKE ?
			`, like, like, "%"+digits+"%")
		} else {
			base = base.Where(`
				LOWER(name) LIKE ? OR
				LOWER(comment) LIKE ?
			`, like, like)
		}
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := base.Order("created_at desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset((f.Page - 1) * f.Limit)
	}

	var recs []inquiryRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*Inquiry, len(recs))
	for i, rec := range recs {
		items[i] = toEntity(rec)
	}
	return items, int(total), nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uint, status Status) error {
	res := r.db.WithContext(ctx).Model(&inquiryRecord{}).
		Where("id = ?", id).
		Update("status", string(status))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toDB(item *Inquiry) inquiryRecord {
	status := item.Status
	if status == "" {
		status = StatusNew
	}
	return inquiryRecord{
		ID:              item.ID,
		Name:            item.Name,
		Phone:           item.Phone,
		PhoneDigits:     item.PhoneDigits,
		Comment:         item.Comment,
		Source:          item.Source,
		PageURL:         item.PageURL,
		ConsentAccepted: item.ConsentAccepted,
		Status:          string(status),
	}
}

func toEntity(rec inquiryRecord) *Inquiry {
	return &Inquiry{
		ID:              rec.ID,
		Name:            rec.Name,
		Phone:           rec.Phone,
		PhoneDigits:     rec.PhoneDigits,
		Comment:         rec.Comment,
		Source:          rec.Source,
		PageURL:         rec.PageURL,
		ConsentAccepted: rec.ConsentAccepted,
		Status:          Status(rec.Status),
		CreatedAt:       rec.CreatedAt,
		UpdatedAt:       rec.UpdatedAt,
	}
}
