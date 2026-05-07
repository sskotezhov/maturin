package banner

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Get(ctx context.Context) ([]BannerItem, error)
	Set(ctx context.Context, items []BannerItem) error
}

type bannerRecord struct {
	ID        uint   `gorm:"primaryKey"`
	ProductID string `gorm:"not null"`
	Position  int    `gorm:"not null;index"`
	UpdatedAt time.Time
}

func (bannerRecord) TableName() string { return "banner_products" }

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&bannerRecord{})
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context) ([]BannerItem, error) {
	var recs []bannerRecord
	if err := r.db.WithContext(ctx).Order("position asc").Find(&recs).Error; err != nil {
		return nil, err
	}
	items := make([]BannerItem, len(recs))
	for i, rec := range recs {
		items[i] = toEntity(rec)
	}
	return items, nil
}

func (r *repository) Set(ctx context.Context, items []BannerItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&bannerRecord{}).Error; err != nil {
			return err
		}
		recs := make([]bannerRecord, len(items))
		for i, item := range items {
			recs[i] = bannerRecord{
				ProductID: item.ProductID,
				Position:  item.Position,
			}
		}
		return tx.Create(&recs).Error
	})
}

func toEntity(rec bannerRecord) BannerItem {
	return BannerItem{
		ID:        rec.ID,
		ProductID: rec.ProductID,
		Position:  rec.Position,
		UpdatedAt: rec.UpdatedAt,
	}
}
