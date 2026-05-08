package inquiry_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/sskotezhov/maturin/internal/inquiry"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

type mockInquiryRepo struct{ mock.Mock }

func (m *mockInquiryRepo) Create(ctx context.Context, item *inquiry.Inquiry) error {
	return m.Called(ctx, item).Error(0)
}

func (m *mockInquiryRepo) FindByID(ctx context.Context, id uint) (*inquiry.Inquiry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inquiry.Inquiry), args.Error(1)
}

func (m *mockInquiryRepo) FindFiltered(ctx context.Context, f inquiry.Filter) ([]*inquiry.Inquiry, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*inquiry.Inquiry), args.Int(1), args.Error(2)
}

func (m *mockInquiryRepo) UpdateStatus(ctx context.Context, id uint, status inquiry.Status) error {
	return m.Called(ctx, id, status).Error(0)
}

type mockInquiryUserRepo struct{ mock.Mock }

func (m *mockInquiryUserRepo) FindByID(ctx context.Context, id uint) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockInquiryUserRepo) FindByIDs(ctx context.Context, ids []uint) ([]*user.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockInquiryUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockInquiryUserRepo) FindAllByRole(ctx context.Context, role roles.Role) ([]*user.User, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockInquiryUserRepo) FindFiltered(ctx context.Context, f user.Filter) ([]*user.User, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Int(1), args.Error(2)
}

func (m *mockInquiryUserRepo) Create(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockInquiryUserRepo) Update(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

type mockInquiryEmailSender struct{ mock.Mock }

func (m *mockInquiryEmailSender) SendInquirySubmitted(to string, id uint, name, phone, comment string) error {
	return m.Called(to, id, name, phone, comment).Error(0)
}
