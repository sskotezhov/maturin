package user

import (
	"context"
	"errors"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrForbiddenField = errors.New("field cannot be updated via profile")
)

var allowedFields = map[string]bool{
	"last_name":    true,
	"first_name":   true,
	"middle_name":  true,
	"phone":        true,
	"telegram":     true,
	"company_name": true,
	"inn":          true,
}

type UpdateInput struct {
	Mask   []string
	Values map[string]string
}

type Service interface {
	GetProfile(ctx context.Context, userID uint) (*User, error)
	UpdateProfile(ctx context.Context, userID uint, input UpdateInput) (*User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProfile(ctx context.Context, userID uint) (*User, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *service) UpdateProfile(ctx context.Context, userID uint, input UpdateInput) (*User, error) {
	for _, field := range input.Mask {
		if !allowedFields[field] {
			return nil, ErrForbiddenField
		}
	}

	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}

	for _, field := range input.Mask {
		val := input.Values[field]
		switch field {
		case "last_name":
			u.LastName = val
		case "first_name":
			u.FirstName = val
		case "middle_name":
			u.MiddleName = val
		case "phone":
			u.Phone = val
		case "telegram":
			u.Telegram = val
		case "company_name":
			u.CompanyName = val
		case "inn":
			u.INN = val
		}
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
