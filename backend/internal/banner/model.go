package banner

import "time"

type BannerItem struct {
	ID        uint
	ProductID string
	Position  int
	UpdatedAt time.Time
}
