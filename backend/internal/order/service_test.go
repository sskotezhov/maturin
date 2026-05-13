package order_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/sskotezhov/maturin/internal/order"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

func newTestOrderService(t *testing.T) (order.Service, *mockOrderRepo, *mockOrderUserRepo, *mockOrderEmailSender, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	orderRepo := &mockOrderRepo{}
	userRepo := &mockOrderUserRepo{}
	emailSender := &mockOrderEmailSender{}
	svc := order.NewService(orderRepo, userRepo, emailSender, rdb, nil)
	return svc, orderRepo, userRepo, emailSender, mr
}

func ptr(f float64) *float64 { return &f }

// AddItem

func TestAddItem_NilPriceAndNoComment_ReturnsError(t *testing.T) {
	svc, _, _, _, _ := newTestOrderService(t)

	_, err := svc.AddItem(context.Background(), 1, order.AddItemInput{
		ProductID:     "p1",
		ProductName:   "Product",
		Quantity:      1,
		PriceSnapshot: nil,
		Comment:       "",
	})
	assert.ErrorIs(t, err, order.ErrCommentRequired)
}

func TestAddItem_NoExistingDraft_CreatesNewOrder(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)

	orderRepo.On("FindDraftByUser", mock.Anything, uint(10)).Return(nil, gorm.ErrRecordNotFound)
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*order.Order")).
		Run(func(args mock.Arguments) {
			o := args.Get(1).(*order.Order)
			o.ID = 99
		}).Return(nil)
	orderRepo.On("AddItem", mock.Anything, mock.AnythingOfType("*order.Item")).Return(nil)
	created := &order.Order{ID: 99, UserID: 10, Status: order.StatusDraft, Items: []order.Item{
		{ID: 1, ProductID: "p1"},
	}}
	orderRepo.On("FindByID", mock.Anything, uint(99)).Return(created, nil)

	result, err := svc.AddItem(context.Background(), 10, order.AddItemInput{
		ProductID:     "p1",
		ProductName:   "Product",
		Quantity:      2,
		PriceSnapshot: ptr(100.0),
	})
	require.NoError(t, err)
	assert.Equal(t, uint(99), result.ID)
	orderRepo.AssertExpectations(t)
}

func TestAddItem_ExistingDraft_AddsItem(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)

	existing := &order.Order{ID: 5, UserID: 10, Status: order.StatusDraft}
	orderRepo.On("FindDraftByUser", mock.Anything, uint(10)).Return(existing, nil)
	orderRepo.On("AddItem", mock.Anything, mock.AnythingOfType("*order.Item")).Return(nil)
	updated := &order.Order{ID: 5, UserID: 10, Status: order.StatusDraft, Items: []order.Item{{ID: 1}}}
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(updated, nil)

	result, err := svc.AddItem(context.Background(), 10, order.AddItemInput{
		ProductID:     "p2",
		ProductName:   "Product 2",
		Quantity:      1,
		PriceSnapshot: ptr(50.0),
	})
	require.NoError(t, err)
	assert.Equal(t, uint(5), result.ID)
	orderRepo.AssertExpectations(t)
}

// Submit

func TestSubmit_NoDraft_ReturnsNotFound(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	orderRepo.On("FindDraftByUser", mock.Anything, uint(1)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.Submit(context.Background(), 1)
	assert.ErrorIs(t, err, order.ErrNotFound)
	orderRepo.AssertExpectations(t)
}

func TestSubmit_Success(t *testing.T) {
	svc, orderRepo, userRepo, _, _ := newTestOrderService(t)

	draft := &order.Order{ID: 3, UserID: 10, Status: order.StatusDraft}
	orderRepo.On("FindDraftByUser", mock.Anything, uint(10)).Return(draft, nil)
	orderRepo.On("UpdateStatus", mock.Anything, uint(3), order.StatusSubmitted, (*float64)(nil)).Return(nil)
	userRepo.On("FindAllByRole", mock.Anything, roles.RoleManager).Return([]*user.User{}, nil)
	userRepo.On("FindAllByRole", mock.Anything, roles.RoleAdmin).Return([]*user.User{}, nil)

	result, err := svc.Submit(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, order.StatusSubmitted, result.Status)
	assert.Equal(t, order.ResponseWaitingManager, result.ResponseStatus)
	orderRepo.AssertExpectations(t)
}

// CancelOrder

func TestCancelOrder_ManagerRole_Forbidden(t *testing.T) {
	svc, _, _, _, _ := newTestOrderService(t)

	err := svc.CancelOrder(context.Background(), 1, string(roles.RoleManager), 10)
	assert.ErrorIs(t, err, order.ErrForbidden)
}

func TestCancelOrder_NotOwner_Forbidden(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 10, UserID: 99, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(10)).Return(ord, nil)

	err := svc.CancelOrder(context.Background(), 1, string(roles.RoleClient), 10)
	assert.ErrorIs(t, err, order.ErrForbidden)
	orderRepo.AssertExpectations(t)
}

func TestCancelOrder_AlreadyApproved_InvalidStatus(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 10, UserID: 1, Status: order.StatusApproved}
	orderRepo.On("FindByID", mock.Anything, uint(10)).Return(ord, nil)

	err := svc.CancelOrder(context.Background(), 1, string(roles.RoleClient), 10)
	assert.ErrorIs(t, err, order.ErrInvalidStatus)
	orderRepo.AssertExpectations(t)
}

func TestCancelOrder_Success(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 10, UserID: 1, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(10)).Return(ord, nil)
	orderRepo.On("UpdateStatus", mock.Anything, uint(10), order.StatusCancelled, (*float64)(nil)).Return(nil)

	err := svc.CancelOrder(context.Background(), 1, string(roles.RoleClient), 10)
	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

// ApproveOrder

func TestApproveOrder_ClientRole_Forbidden(t *testing.T) {
	svc, _, _, _, _ := newTestOrderService(t)

	_, err := svc.ApproveOrder(context.Background(), 1, string(roles.RoleClient), 5, 999.0)
	assert.ErrorIs(t, err, order.ErrForbidden)
}

func TestApproveOrder_NotSubmitted_InvalidStatus(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 5, UserID: 1, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(ord, nil)

	_, err := svc.ApproveOrder(context.Background(), 99, string(roles.RoleManager), 5, 999.0)
	assert.ErrorIs(t, err, order.ErrInvalidStatus)
	orderRepo.AssertExpectations(t)
}

func TestApproveOrder_Success(t *testing.T) {
	svc, orderRepo, userRepo, emailSender, _ := newTestOrderService(t)
	ord := &order.Order{ID: 5, UserID: 10, Status: order.StatusSubmitted}
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(ord, nil)
	orderRepo.On("UpdateStatus", mock.Anything, uint(5), order.StatusApproved, mock.AnythingOfType("*float64")).Return(nil)
	client := &user.User{ID: 10, Email: "client@example.com", Role: roles.RoleClient}
	userRepo.On("FindByID", mock.Anything, uint(10)).Return(client, nil)
	emailSender.On("SendOrderApproved", "client@example.com", uint(5), 1500.0).Return(nil)

	result, err := svc.ApproveOrder(context.Background(), 99, string(roles.RoleManager), 5, 1500.0)
	require.NoError(t, err)
	assert.Equal(t, order.StatusApproved, result.Status)
	require.NotNil(t, result.TotalPrice)
	assert.InDelta(t, 1500.0, *result.TotalPrice, 0.001)
	orderRepo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

// GetOrder

func TestGetOrder_ClientAccessesOwnDraftOrder(t *testing.T) {
	svc, orderRepo, userRepo, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 10, Status: order.StatusDraft, Items: []order.Item{}}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)
	userRepo.On("FindByID", mock.Anything, uint(10)).Return(&user.User{ID: 10}, nil)

	result, err := svc.GetOrder(context.Background(), 10, string(roles.RoleClient), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
	orderRepo.AssertExpectations(t)
}

func TestGetOrder_ClientAccessesOtherOrder_Forbidden(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 99, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)

	_, err := svc.GetOrder(context.Background(), 10, string(roles.RoleClient), 1)
	assert.ErrorIs(t, err, order.ErrForbidden)
	orderRepo.AssertExpectations(t)
}

func TestGetOrder_ManagerAccessesForeignDraft_NotFound(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 99, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)

	_, err := svc.GetOrder(context.Background(), 10, string(roles.RoleManager), 1)
	assert.ErrorIs(t, err, order.ErrNotFound)
	orderRepo.AssertExpectations(t)
}

func TestGetOrder_ManagerAccessesOwnDraft(t *testing.T) {
	svc, orderRepo, userRepo, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 10, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)
	userRepo.On("FindByID", mock.Anything, uint(10)).Return(&user.User{ID: 10}, nil)

	result, err := svc.GetOrder(context.Background(), 10, string(roles.RoleManager), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
	orderRepo.AssertExpectations(t)
}

// SendMessage

func TestSendMessage_DraftStatus_InvalidStatus(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 10, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)

	_, err := svc.SendMessage(context.Background(), 10, string(roles.RoleClient), 1, "hello")
	assert.ErrorIs(t, err, order.ErrInvalidStatus)
	orderRepo.AssertExpectations(t)
}

func TestSendMessage_ClientAccessesOtherOrder_Forbidden(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 99, Status: order.StatusSubmitted}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)

	_, err := svc.SendMessage(context.Background(), 10, string(roles.RoleClient), 1, "hello")
	assert.ErrorIs(t, err, order.ErrForbidden)
	orderRepo.AssertExpectations(t)
}

func TestSendMessage_ManagerSendsToSubmittedOrder(t *testing.T) {
	svc, orderRepo, userRepo, emailSender, _ := newTestOrderService(t)
	ord := &order.Order{ID: 1, UserID: 10, Status: order.StatusSubmitted}
	orderRepo.On("FindByID", mock.Anything, uint(1)).Return(ord, nil)
	orderRepo.On("AddMessage", mock.Anything, mock.AnythingOfType("*order.Message")).Return(nil)
	client := &user.User{ID: 10, Email: "client@example.com", Role: roles.RoleClient}
	userRepo.On("FindByID", mock.Anything, uint(10)).Return(client, nil)
	emailSender.On("SendNewMessage", "client@example.com", uint(1)).Return(nil)

	msg, err := svc.SendMessage(context.Background(), 99, string(roles.RoleManager), 1, "Your order is ready")
	require.NoError(t, err)
	assert.Equal(t, "Your order is ready", msg.Text)
	orderRepo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

// StaffAddItem

func TestStaffAddItem_ApprovedStatus_NotEditable(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	ord := &order.Order{ID: 5, UserID: 10, Status: order.StatusApproved}
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(ord, nil)

	_, err := svc.StaffAddItem(context.Background(), 99, 5, order.AddItemInput{
		ProductID:     "p1",
		ProductName:   "Product",
		Quantity:      1,
		PriceSnapshot: ptr(100.0),
	})
	assert.ErrorIs(t, err, order.ErrOrderNotEditable)
	orderRepo.AssertExpectations(t)
}

func TestStaffAddItem_DraftStatus_Success(t *testing.T) {
	svc, orderRepo, userRepo, emailSender, _ := newTestOrderService(t)
	ord := &order.Order{ID: 5, UserID: 10, Status: order.StatusDraft}
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(ord, nil)
	orderRepo.On("AddItem", mock.Anything, mock.AnythingOfType("*order.Item")).Return(nil)
	updated := &order.Order{ID: 5, UserID: 10, Status: order.StatusDraft, Items: []order.Item{{ID: 1}}}
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(updated, nil)
	client := &user.User{ID: 10, Email: "client@example.com"}
	userRepo.On("FindByID", mock.Anything, uint(10)).Return(client, nil)
	emailSender.On("SendOrderModified", "client@example.com", uint(5)).Return(nil)

	// Staff actor (99) != order owner (10) → notification sent
	result, err := svc.StaffAddItem(context.Background(), 99, 5, order.AddItemInput{
		ProductID:     "p1",
		ProductName:   "Product",
		Quantity:      3,
		PriceSnapshot: ptr(200.0),
	})
	require.NoError(t, err)
	assert.Equal(t, uint(5), result.ID)

	// Allow some time for the async-ish notification (it's actually sync in tryNotify path)
	time.Sleep(10 * time.Millisecond)
	orderRepo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

// UpdateItem

func TestUpdateItem_NotOwner_Forbidden(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	item := &order.Item{ID: 1, OrderID: 5}
	ord := &order.Order{ID: 5, UserID: 99, Status: order.StatusDraft}
	orderRepo.On("FindItem", mock.Anything, uint(1)).Return(item, nil)
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(ord, nil)

	_, err := svc.UpdateItem(context.Background(), 10, 1, order.UpdateItemInput{Quantity: 2})
	assert.ErrorIs(t, err, order.ErrForbidden)
	orderRepo.AssertExpectations(t)
}

func TestUpdateItem_SubmittedStatus_InvalidStatus(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	item := &order.Item{ID: 1, OrderID: 5}
	ord := &order.Order{ID: 5, UserID: 10, Status: order.StatusSubmitted}
	orderRepo.On("FindItem", mock.Anything, uint(1)).Return(item, nil)
	orderRepo.On("FindByID", mock.Anything, uint(5)).Return(ord, nil)

	_, err := svc.UpdateItem(context.Background(), 10, 1, order.UpdateItemInput{Quantity: 2})
	assert.ErrorIs(t, err, order.ErrInvalidStatus)
	orderRepo.AssertExpectations(t)
}

// DeleteItem

func TestDeleteItem_NotOwner_Forbidden(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	item := &order.Item{ID: 2, OrderID: 7}
	ord := &order.Order{ID: 7, UserID: 99, Status: order.StatusDraft}
	orderRepo.On("FindItem", mock.Anything, uint(2)).Return(item, nil)
	orderRepo.On("FindByID", mock.Anything, uint(7)).Return(ord, nil)

	err := svc.DeleteItem(context.Background(), 10, 2)
	assert.ErrorIs(t, err, order.ErrForbidden)
	orderRepo.AssertExpectations(t)
}

func TestDeleteItem_Success(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	item := &order.Item{ID: 2, OrderID: 7}
	ord := &order.Order{ID: 7, UserID: 10, Status: order.StatusDraft}
	orderRepo.On("FindItem", mock.Anything, uint(2)).Return(item, nil)
	orderRepo.On("FindByID", mock.Anything, uint(7)).Return(ord, nil)
	orderRepo.On("DeleteItem", mock.Anything, uint(2)).Return(nil)

	err := svc.DeleteItem(context.Background(), 10, 2)
	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

// GetCart

func TestGetCart_NoDraft_NotFound(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	orderRepo.On("FindDraftByUser", mock.Anything, uint(1)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.GetCart(context.Background(), 1)
	assert.ErrorIs(t, err, order.ErrNotFound)
	orderRepo.AssertExpectations(t)
}

func TestGetCart_ReturnsDraft(t *testing.T) {
	svc, orderRepo, _, _, _ := newTestOrderService(t)
	draft := &order.Order{ID: 3, UserID: 1, Status: order.StatusDraft}
	orderRepo.On("FindDraftByUser", mock.Anything, uint(1)).Return(draft, nil)

	result, err := svc.GetCart(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(3), result.ID)
	assert.Equal(t, order.ResponseNone, result.ResponseStatus)
	orderRepo.AssertExpectations(t)
}

// Utility — ensure errors package is used
var _ = errors.New
