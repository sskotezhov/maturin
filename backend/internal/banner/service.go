package banner

import (
	"context"
	"errors"
)

var (
	ErrInvalidCount     = errors.New("banner must contain exactly 4 products")
	ErrDuplicateProduct = errors.New("banner product IDs must be unique")
)

const BannerSize = 4

type Service interface {
	Get(ctx context.Context) ([]BannerItem, error)
	Set(ctx context.Context, productIDs []string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Get(ctx context.Context) ([]BannerItem, error) {
	return s.repo.Get(ctx)
}

func (s *service) Set(ctx context.Context, productIDs []string) error {
	if len(productIDs) != BannerSize {
		return ErrInvalidCount
	}

	seen := make(map[string]struct{}, BannerSize)
	for _, id := range productIDs {
		if _, exists := seen[id]; exists {
			return ErrDuplicateProduct
		}
		seen[id] = struct{}{}
	}

	items := make([]BannerItem, BannerSize)
	for i, id := range productIDs {
		items[i] = BannerItem{
			ProductID: id,
			Position:  i + 1,
		}
	}
	return s.repo.Set(ctx, items)
}
