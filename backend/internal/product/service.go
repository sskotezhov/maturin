package product

import (
	"context"
	"errors"
	"sort"
	"strings"
)

var ErrNotFound = errors.New("product not found")

type ListFilter struct {
	CategoryKey string
	Type        string
	InStock     *bool
	HasPrice    *bool
	MinPrice    *float64
	MaxPrice    *float64
	Search      string
	SortBy      string
	SortDir     string
	Page        int
	Limit       int
}

type ListResult struct {
	Items []Product `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type Service interface {
	ListProducts(ctx context.Context, f ListFilter) (ListResult, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	ListCategories(ctx context.Context) ([]Category, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListProducts(ctx context.Context, f ListFilter) (ListResult, error) {
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return ListResult{}, err
	}

	search := strings.ToLower(strings.TrimSpace(f.Search))

	filtered := make([]Product, 0, len(all))
	for _, p := range all {
		if f.CategoryKey != "" && p.CategoryKey != f.CategoryKey {
			continue
		}
		if f.Type != "" && p.Type != f.Type {
			continue
		}
		if f.InStock != nil && p.InStock != *f.InStock {
			continue
		}
		if f.HasPrice != nil {
			hasPrice := p.Price != nil
			if hasPrice != *f.HasPrice {
				continue
			}
		}
		if f.MinPrice != nil && (p.Price == nil || *p.Price < *f.MinPrice) {
			continue
		}
		if f.MaxPrice != nil && (p.Price == nil || *p.Price > *f.MaxPrice) {
			continue
		}
		if search != "" {
			if !matchSearch(p, search) {
				continue
			}
		}
		filtered = append(filtered, p)
	}

	sortProducts(filtered, f.SortBy, f.SortDir)

	total := len(filtered)

	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	start := (f.Page - 1) * f.Limit
	if start >= total {
		return ListResult{Items: []Product{}, Total: total, Page: f.Page, Limit: f.Limit}, nil
	}
	end := min(start+f.Limit, total)

	return ListResult{
		Items: filtered[start:end],
		Total: total,
		Page:  f.Page,
		Limit: f.Limit,
	}, nil
}

func (s *service) GetProduct(ctx context.Context, id string) (*Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.GetCategories(ctx)
}

func matchSearch(p Product, q string) bool {
	return strings.Contains(strings.ToLower(p.Name), q) ||
		strings.Contains(strings.ToLower(p.FullName), q) ||
		strings.Contains(strings.ToLower(p.Article), q) ||
		strings.Contains(strings.ToLower(p.Code), q)
}

func sortProducts(items []Product, by, dir string) {
	desc := strings.ToLower(dir) == "desc"

	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		var less bool
		switch by {
		case "price":
			ap := priceVal(a.Price)
			bp := priceVal(b.Price)
			less = ap < bp
		case "code":
			less = a.Code < b.Code
		case "updated_at":
			less = a.UpdatedAt < b.UpdatedAt
		default: // name
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
		if desc {
			return !less
		}
		return less
	})
}

func priceVal(p *float64) float64 {
	if p == nil {
		return -1
	}
	return *p
}
