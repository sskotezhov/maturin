package order_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/sskotezhov/maturin/internal/order"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

type mockOrderRepo struct{ mock.Mock }

func (m *mockOrderRepo) FindDraftByUser(ctx context.Context, userID uint) (*order.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *mockOrderRepo) FindByID(ctx context.Context, id uint) (*order.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *mockOrderRepo) FindFiltered(ctx context.Context, f order.Filter) ([]*order.Order, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*order.Order), args.Error(1)
}

func (m *mockOrderRepo) FindByUserID(ctx context.Context, userID uint, limit int) ([]*order.Order, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*order.Order), args.Error(1)
}

func (m *mockOrderRepo) CountByUserID(ctx context.Context, userID uint) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockOrderRepo) CountByStatus(ctx context.Context) (map[order.Status]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[order.Status]int), args.Error(1)
}

func (m *mockOrderRepo) CountStaleSubmitted(ctx context.Context, threshold time.Duration) (int, error) {
	args := m.Called(ctx, threshold)
	return args.Int(0), args.Error(1)
}

func (m *mockOrderRepo) Create(ctx context.Context, o *order.Order) error {
	return m.Called(ctx, o).Error(0)
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uint, status order.Status, totalPrice *float64) error {
	return m.Called(ctx, id, status, totalPrice).Error(0)
}

func (m *mockOrderRepo) AddItem(ctx context.Context, item *order.Item) error {
	return m.Called(ctx, item).Error(0)
}

func (m *mockOrderRepo) UpdateItem(ctx context.Context, item *order.Item) error {
	return m.Called(ctx, item).Error(0)
}

func (m *mockOrderRepo) DeleteItem(ctx context.Context, itemID uint) error {
	return m.Called(ctx, itemID).Error(0)
}

func (m *mockOrderRepo) FindItem(ctx context.Context, itemID uint) (*order.Item, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Item), args.Error(1)
}

func (m *mockOrderRepo) AddMessage(ctx context.Context, msg *order.Message) error {
	return m.Called(ctx, msg).Error(0)
}

func (m *mockOrderRepo) FindMessages(ctx context.Context, orderID uint) ([]order.Message, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.Message), args.Error(1)
}

func (m *mockOrderRepo) FindLastMessage(ctx context.Context, orderID uint) (*order.Message, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Message), args.Error(1)
}

func (m *mockOrderRepo) SetOnecRef(ctx context.Context, id uint, ref, number string) error {
	args := m.Called(ctx, id, ref, number)
	return args.Error(0)
}

type mockOrderUserRepo struct{ mock.Mock }

func (m *mockOrderUserRepo) FindByID(ctx context.Context, id uint) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockOrderUserRepo) FindByIDs(ctx context.Context, ids []uint) ([]*user.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockOrderUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockOrderUserRepo) FindAllByRole(ctx context.Context, role roles.Role) ([]*user.User, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockOrderUserRepo) FindFiltered(ctx context.Context, f user.Filter) ([]*user.User, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Int(1), args.Error(2)
}

func (m *mockOrderUserRepo) Create(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockOrderUserRepo) Update(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

type mockOrderEmailSender struct{ mock.Mock }

func (m *mockOrderEmailSender) SendOrderSubmitted(to string, orderID uint) error {
	return m.Called(to, orderID).Error(0)
}

func (m *mockOrderEmailSender) SendOrderApproved(to string, orderID uint, totalPrice float64) error {
	return m.Called(to, orderID, totalPrice).Error(0)
}

func (m *mockOrderEmailSender) SendNewMessage(to string, orderID uint) error {
	return m.Called(to, orderID).Error(0)
}

func (m *mockOrderEmailSender) SendOrderModified(to string, orderID uint) error {
	return m.Called(to, orderID).Error(0)
}
