package auth_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) FindByIDs(ctx context.Context, ids []uint) ([]*user.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) FindAllByRole(ctx context.Context, role roles.Role) ([]*user.User, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockUserRepo) FindFiltered(ctx context.Context, f user.Filter) ([]*user.User, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Int(1), args.Error(2)
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

type mockAuthEmailSender struct{ mock.Mock }

func (m *mockAuthEmailSender) SendVerificationCode(to, code string) error {
	return m.Called(to, code).Error(0)
}

func (m *mockAuthEmailSender) SendPasswordResetCode(to, code string) error {
	return m.Called(to, code).Error(0)
}
