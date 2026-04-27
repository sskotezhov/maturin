package order

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Filter struct {
	Status string
	UserID uint
	Page   int
	Limit  int
}

type Repository interface {
	FindDraftByUser(ctx context.Context, userID uint) (*Order, error)
	FindByID(ctx context.Context, id uint) (*Order, error)
	FindFiltered(ctx context.Context, f Filter) ([]*Order, error)
	FindByUserID(ctx context.Context, userID uint, limit int) ([]*Order, error)
	CountByUserID(ctx context.Context, userID uint) (int, error)
	CountByStatus(ctx context.Context) (map[Status]int, error)
	CountStaleSubmitted(ctx context.Context, threshold time.Duration) (int, error)
	Create(ctx context.Context, order *Order) error
	UpdateStatus(ctx context.Context, id uint, status Status, totalPrice *float64) error
	AddItem(ctx context.Context, item *Item) error
	UpdateItem(ctx context.Context, item *Item) error
	DeleteItem(ctx context.Context, itemID uint) error
	FindItem(ctx context.Context, itemID uint) (*Item, error)
	AddMessage(ctx context.Context, msg *Message) error
	FindMessages(ctx context.Context, orderID uint) ([]Message, error)
}

type orderRecord struct {
	ID         uint         `gorm:"primaryKey"`
	UserID     uint         `gorm:"not null;index"`
	Status     string       `gorm:"type:varchar(20);not null;default:'draft'"`
	TotalPrice *float64     `gorm:"type:numeric(12,2)"`
	Items      []itemRecord `gorm:"foreignKey:OrderID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (orderRecord) TableName() string { return "orders" }

type itemRecord struct {
	ID            uint     `gorm:"primaryKey"`
	OrderID       uint     `gorm:"not null;index"`
	ProductID     string   `gorm:"not null"`
	ProductName   string   `gorm:"not null"`
	ProductCode   string   `gorm:"not null;default:''"`
	Quantity      int      `gorm:"not null"`
	PriceSnapshot *float64 `gorm:"type:numeric(12,2)"`
	Comment       string   `gorm:"not null;default:''"`
	CreatedAt     time.Time
}

func (itemRecord) TableName() string { return "request_items" }

type messageRecord struct {
	ID        uint   `gorm:"primaryKey"`
	OrderID   uint   `gorm:"not null;index"`
	UserID    uint   `gorm:"not null"`
	Text      string `gorm:"not null"`
	CreatedAt time.Time
}

func (messageRecord) TableName() string { return "request_messages" }

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&orderRecord{}, &itemRecord{}, &messageRecord{})
}

func toOrderEntity(rec orderRecord) *Order {
	items := make([]Item, len(rec.Items))
	for i, item := range rec.Items {
		items[i] = toItemEntity(item)
	}
	return &Order{
		ID:         rec.ID,
		UserID:     rec.UserID,
		Status:     Status(rec.Status),
		TotalPrice: rec.TotalPrice,
		Items:      items,
		CreatedAt:  rec.CreatedAt,
		UpdatedAt:  rec.UpdatedAt,
	}
}

func toItemEntity(rec itemRecord) Item {
	return Item{
		ID:            rec.ID,
		OrderID:       rec.OrderID,
		ProductID:     rec.ProductID,
		ProductName:   rec.ProductName,
		ProductCode:   rec.ProductCode,
		Quantity:      rec.Quantity,
		PriceSnapshot: rec.PriceSnapshot,
		Comment:       rec.Comment,
		CreatedAt:     rec.CreatedAt,
	}
}

func toMessageEntity(rec messageRecord) Message {
	return Message{
		ID:        rec.ID,
		OrderID:   rec.OrderID,
		UserID:    rec.UserID,
		Text:      rec.Text,
		CreatedAt: rec.CreatedAt,
	}
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindDraftByUser(ctx context.Context, userID uint) (*Order, error) {
	var rec orderRecord
	err := r.db.WithContext(ctx).Preload("Items").
		Where("user_id = ? AND status = ?", userID, string(StatusDraft)).
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return toOrderEntity(rec), nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*Order, error) {
	var rec orderRecord
	if err := r.db.WithContext(ctx).Preload("Items").First(&rec, id).Error; err != nil {
		return nil, err
	}
	return toOrderEntity(rec), nil
}

func (r *repository) FindFiltered(ctx context.Context, f Filter) ([]*Order, error) {
	q := r.db.WithContext(ctx).Preload("Items").Order("created_at desc")
	if f.UserID != 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset((f.Page - 1) * f.Limit)
	}
	var recs []orderRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, err
	}
	orders := make([]*Order, len(recs))
	for i, rec := range recs {
		orders[i] = toOrderEntity(rec)
	}
	return orders, nil
}

func (r *repository) Create(ctx context.Context, order *Order) error {
	rec := orderRecord{
		UserID: order.UserID,
		Status: string(order.Status),
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	order.ID = rec.ID
	order.CreatedAt = rec.CreatedAt
	order.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uint, status Status, totalPrice *float64) error {
	updates := map[string]any{"status": string(status)}
	if totalPrice != nil {
		updates["total_price"] = totalPrice
	}
	return r.db.WithContext(ctx).Model(&orderRecord{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) AddItem(ctx context.Context, item *Item) error {
	rec := itemRecord{
		OrderID:       item.OrderID,
		ProductID:     item.ProductID,
		ProductName:   item.ProductName,
		ProductCode:   item.ProductCode,
		Quantity:      item.Quantity,
		PriceSnapshot: item.PriceSnapshot,
		Comment:       item.Comment,
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	item.ID = rec.ID
	item.CreatedAt = rec.CreatedAt
	return nil
}

func (r *repository) UpdateItem(ctx context.Context, item *Item) error {
	return r.db.WithContext(ctx).Model(&itemRecord{}).Where("id = ?", item.ID).
		Updates(map[string]any{
			"quantity": item.Quantity,
			"comment":  item.Comment,
		}).Error
}

func (r *repository) DeleteItem(ctx context.Context, itemID uint) error {
	return r.db.WithContext(ctx).Delete(&itemRecord{}, itemID).Error
}

func (r *repository) FindItem(ctx context.Context, itemID uint) (*Item, error) {
	var rec itemRecord
	if err := r.db.WithContext(ctx).First(&rec, itemID).Error; err != nil {
		return nil, err
	}
	item := toItemEntity(rec)
	return &item, nil
}

func (r *repository) AddMessage(ctx context.Context, msg *Message) error {
	rec := messageRecord{
		OrderID: msg.OrderID,
		UserID:  msg.UserID,
		Text:    msg.Text,
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	msg.ID = rec.ID
	msg.CreatedAt = rec.CreatedAt
	return nil
}

func (r *repository) FindMessages(ctx context.Context, orderID uint) ([]Message, error) {
	var recs []messageRecord
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at asc").Find(&recs).Error; err != nil {
		return nil, err
	}
	msgs := make([]Message, len(recs))
	for i, rec := range recs {
		msgs[i] = toMessageEntity(rec)
	}
	return msgs, nil
}

func (r *repository) FindByUserID(ctx context.Context, userID uint, limit int) ([]*Order, error) {
	q := r.db.WithContext(ctx).Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at desc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var recs []orderRecord
	if err := q.Find(&recs).Error; err != nil {
		return nil, err
	}
	orders := make([]*Order, len(recs))
	for i, rec := range recs {
		orders[i] = toOrderEntity(rec)
	}
	return orders, nil
}

func (r *repository) CountByUserID(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&orderRecord{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

func (r *repository) CountByStatus(ctx context.Context) (map[Status]int, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).Model(&orderRecord{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[Status]int{
		StatusDraft:     0,
		StatusSubmitted: 0,
		StatusApproved:  0,
		StatusCancelled: 0,
	}
	for _, r := range rows {
		out[Status(r.Status)] = int(r.Count)
	}
	return out, nil
}

func (r *repository) CountStaleSubmitted(ctx context.Context, threshold time.Duration) (int, error) {
	cutoff := time.Now().Add(-threshold)
	var count int64
	err := r.db.WithContext(ctx).Model(&orderRecord{}).
		Where("status = ? AND updated_at < ?", string(StatusSubmitted), cutoff).
		Count(&count).Error
	return int(count), err
}
