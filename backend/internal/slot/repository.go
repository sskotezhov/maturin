package slot

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AvailableFilter struct {
	Date      *time.Time
	ManagerID *uint
}

type Filter struct {
	Status    string
	Date      *time.Time
	ManagerID *uint
	Page      int
	Limit     int
}

type BookingFilter struct {
	Page  int
	Limit int
}

type Repository interface {
	CreateBatch(ctx context.Context, slots []CallSlot) error
	FindAvailable(ctx context.Context, f AvailableFilter) ([]*CallSlot, error)
	FindByID(ctx context.Context, id uint) (*CallSlot, error)
	Book(ctx context.Context, slotID uint, booking *CallBooking) error
	FindFiltered(ctx context.Context, f Filter) ([]*CallSlot, int, error)
	DeleteByID(ctx context.Context, id uint) error
	FindBookings(ctx context.Context, f BookingFilter) ([]*CallBooking, int, error)
}

type slotRecord struct {
	ID        uint      `gorm:"primaryKey"`
	ManagerID uint      `gorm:"not null;index"`
	StartAt   time.Time `gorm:"not null;index"`
	EndAt     time.Time `gorm:"not null"`
	IsBooked  bool      `gorm:"not null;default:false;index"`
	CreatedAt time.Time
}

func (slotRecord) TableName() string { return "call_slots" }

type bookingRecord struct {
	ID          uint   `gorm:"primaryKey"`
	SlotID      uint   `gorm:"not null;uniqueIndex"`
	Name        string `gorm:"not null"`
	Phone       string `gorm:"not null"`
	PhoneDigits string `gorm:"not null;index"`
	CreatedAt   time.Time
}

func (bookingRecord) TableName() string { return "call_bookings" }

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&slotRecord{}); err != nil {
		return err
	}
	return db.AutoMigrate(&bookingRecord{})
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBatch(ctx context.Context, slots []CallSlot) error {
	recs := make([]slotRecord, len(slots))
	for i, s := range slots {
		recs[i] = slotToRecord(s)
	}
	return r.db.WithContext(ctx).Create(&recs).Error
}

func (r *repository) FindAvailable(ctx context.Context, f AvailableFilter) ([]*CallSlot, error) {
	q := r.db.WithContext(ctx).
		Where("is_booked = false AND start_at > ?", time.Now())

	if f.Date != nil {
		start := time.Date(f.Date.Year(), f.Date.Month(), f.Date.Day(), 0, 0, 0, 0, f.Date.Location())
		end := start.Add(24 * time.Hour)
		q = q.Where("start_at >= ? AND start_at < ?", start, end)
	}
	if f.ManagerID != nil {
		q = q.Where("manager_id = ?", *f.ManagerID)
	}

	var recs []slotRecord
	if err := q.Order("start_at asc").Find(&recs).Error; err != nil {
		return nil, err
	}

	items := make([]*CallSlot, len(recs))
	for i, rec := range recs {
		items[i] = slotToEntity(rec)
	}
	return items, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*CallSlot, error) {
	var rec slotRecord
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		return nil, err
	}
	return slotToEntity(rec), nil
}

func (r *repository) Book(ctx context.Context, slotID uint, booking *CallBooking) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rec slotRecord
		if err := tx.First(&rec, slotID).Error; err != nil {
			return err
		}
		if rec.IsBooked {
			return ErrSlotAlreadyBooked
		}

		if err := tx.Model(&slotRecord{}).Where("id = ?", slotID).Update("is_booked", true).Error; err != nil {
			return err
		}

		bRec := bookingRecord{
			SlotID:      slotID,
			Name:        booking.Name,
			Phone:       booking.Phone,
			PhoneDigits: booking.PhoneDigits,
		}
		if err := tx.Create(&bRec).Error; err != nil {
			return err
		}
		booking.ID = bRec.ID
		booking.SlotID = bRec.SlotID
		booking.CreatedAt = bRec.CreatedAt
		return nil
	})
}

func (r *repository) FindFiltered(ctx context.Context, f Filter) ([]*CallSlot, int, error) {
	base := r.db.WithContext(ctx).Model(&slotRecord{})

	switch f.Status {
	case "free":
		base = base.Where("is_booked = false")
	case "booked":
		base = base.Where("is_booked = true")
	}
	if f.Date != nil {
		start := time.Date(f.Date.Year(), f.Date.Month(), f.Date.Day(), 0, 0, 0, 0, f.Date.Location())
		end := start.Add(24 * time.Hour)
		base = base.Where("start_at >= ? AND start_at < ?", start, end)
	}
	if f.ManagerID != nil {
		base = base.Where("manager_id = ?", *f.ManagerID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := base.Order("start_at asc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset((f.Page - 1) * f.Limit)
	}

	var recs []slotRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*CallSlot, len(recs))
	for i, rec := range recs {
		items[i] = slotToEntity(rec)
	}
	return items, int(total), nil
}

func (r *repository) DeleteByID(ctx context.Context, id uint) error {
	var rec slotRecord
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSlotNotFound
		}
		return err
	}
	if rec.IsBooked {
		return ErrSlotBooked
	}
	return r.db.WithContext(ctx).Delete(&slotRecord{}, id).Error
}

func (r *repository) FindBookings(ctx context.Context, f BookingFilter) ([]*CallBooking, int, error) {
	base := r.db.WithContext(ctx).Model(&bookingRecord{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q := base.Order("created_at desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset((f.Page - 1) * f.Limit)
	}

	var recs []bookingRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*CallBooking, len(recs))
	for i, rec := range recs {
		items[i] = bookingToEntity(rec)
	}
	return items, int(total), nil
}

func slotToRecord(s CallSlot) slotRecord {
	return slotRecord{
		ID:        s.ID,
		ManagerID: s.ManagerID,
		StartAt:   s.StartAt,
		EndAt:     s.EndAt,
		IsBooked:  s.IsBooked,
	}
}

func slotToEntity(rec slotRecord) *CallSlot {
	return &CallSlot{
		ID:        rec.ID,
		ManagerID: rec.ManagerID,
		StartAt:   rec.StartAt,
		EndAt:     rec.EndAt,
		IsBooked:  rec.IsBooked,
		CreatedAt: rec.CreatedAt,
	}
}

func bookingToEntity(rec bookingRecord) *CallBooking {
	return &CallBooking{
		ID:          rec.ID,
		SlotID:      rec.SlotID,
		Name:        rec.Name,
		Phone:       rec.Phone,
		PhoneDigits: rec.PhoneDigits,
		CreatedAt:   rec.CreatedAt,
	}
}
