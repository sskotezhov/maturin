package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidStatus   = errors.New("invalid status for this operation")
	ErrCommentRequired = errors.New("comment required for items without price")
)

type EmailSender interface {
	SendOrderSubmitted(to string, orderID uint) error
	SendOrderApproved(to string, orderID uint, totalPrice float64) error
	SendNewMessage(to string, orderID uint) error
}

type AddItemInput struct {
	ProductID     string
	ProductName   string
	ProductCode   string
	Quantity      int
	PriceSnapshot *float64
	Comment       string
}

type UpdateItemInput struct {
	Quantity int
	Comment  string
}

type Service interface {
	GetCart(ctx context.Context, userID uint) (*Order, error)
	AddItem(ctx context.Context, userID uint, input AddItemInput) (*Order, error)
	UpdateItem(ctx context.Context, userID, itemID uint, input UpdateItemInput) (*Order, error)
	DeleteItem(ctx context.Context, userID, itemID uint) error
	Submit(ctx context.Context, userID uint) (*Order, error)

	GetOrders(ctx context.Context, userID uint, role string, f Filter) ([]*Order, error)
	GetOrder(ctx context.Context, userID uint, role string, orderID uint) (*Order, error)
	CancelOrder(ctx context.Context, userID uint, role string, orderID uint) error
	ApproveOrder(ctx context.Context, userID uint, role string, orderID uint, totalPrice float64) (*Order, error)

	GetMessages(ctx context.Context, userID uint, role string, orderID uint) ([]Message, error)
	SendMessage(ctx context.Context, userID uint, role string, orderID uint, text string) (*Message, error)
}

type service struct {
	repo        Repository
	userRepo    user.Repository
	emailSender EmailSender
	rdb         *redis.Client
}

func NewService(repo Repository, userRepo user.Repository, emailSender EmailSender, rdb *redis.Client) Service {
	return &service{
		repo:        repo,
		userRepo:    userRepo,
		emailSender: emailSender,
		rdb:         rdb,
	}
}

func (s *service) GetCart(ctx context.Context, userID uint) (*Order, error) {
	order, err := s.repo.FindDraftByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		slog.Error("get cart failed", "user_id", userID, "err", err)
		return nil, err
	}
	return order, nil
}

func (s *service) AddItem(ctx context.Context, userID uint, input AddItemInput) (*Order, error) {
	if input.PriceSnapshot == nil && input.Comment == "" {
		return nil, ErrCommentRequired
	}

	order, err := s.repo.FindDraftByUser(ctx, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("find draft failed", "user_id", userID, "err", err)
			return nil, err
		}
		slog.Info("no draft found, creating new order", "user_id", userID)
		order = &Order{UserID: userID, Status: StatusDraft}
		if err := s.repo.Create(ctx, order); err != nil {
			slog.Error("create order failed", "user_id", userID, "err", err)
			return nil, err
		}
		slog.Info("draft order created", "order_id", order.ID, "user_id", userID)
	}

	item := &Item{
		OrderID:       order.ID,
		ProductID:     input.ProductID,
		ProductName:   input.ProductName,
		ProductCode:   input.ProductCode,
		Quantity:      input.Quantity,
		PriceSnapshot: input.PriceSnapshot,
		Comment:       input.Comment,
	}
	if err := s.repo.AddItem(ctx, item); err != nil {
		slog.Error("add item failed", "order_id", order.ID, "product_id", input.ProductID, "err", err)
		return nil, err
	}
	slog.Info("item added to order", "order_id", order.ID, "product_id", input.ProductID, "qty", input.Quantity)

	return s.repo.FindByID(ctx, order.ID)
}

func (s *service) UpdateItem(ctx context.Context, userID, itemID uint, input UpdateItemInput) (*Order, error) {
	item, err := s.repo.FindItem(ctx, itemID)
	if err != nil {
		return nil, ErrNotFound
	}

	order, err := s.repo.FindByID(ctx, item.OrderID)
	if err != nil {
		return nil, ErrNotFound
	}
	if order.UserID != userID {
		slog.Warn("update item forbidden", "user_id", userID, "order_id", order.ID, "owner_id", order.UserID)
		return nil, ErrForbidden
	}
	if order.Status != StatusDraft {
		slog.Warn("update item rejected: order not in draft", "order_id", order.ID, "status", order.Status)
		return nil, ErrInvalidStatus
	}

	item.Quantity = input.Quantity
	item.Comment = input.Comment
	if err := s.repo.UpdateItem(ctx, item); err != nil {
		slog.Error("update item failed", "item_id", itemID, "err", err)
		return nil, err
	}
	slog.Info("item updated", "item_id", itemID, "order_id", order.ID, "qty", input.Quantity)

	return s.repo.FindByID(ctx, order.ID)
}

func (s *service) DeleteItem(ctx context.Context, userID, itemID uint) error {
	item, err := s.repo.FindItem(ctx, itemID)
	if err != nil {
		return ErrNotFound
	}

	order, err := s.repo.FindByID(ctx, item.OrderID)
	if err != nil {
		return ErrNotFound
	}
	if order.UserID != userID {
		slog.Warn("delete item forbidden", "user_id", userID, "order_id", order.ID, "owner_id", order.UserID)
		return ErrForbidden
	}
	if order.Status != StatusDraft {
		slog.Warn("delete item rejected: order not in draft", "order_id", order.ID, "status", order.Status)
		return ErrInvalidStatus
	}

	if err := s.repo.DeleteItem(ctx, itemID); err != nil {
		slog.Error("delete item failed", "item_id", itemID, "err", err)
		return err
	}
	slog.Info("item deleted", "item_id", itemID, "order_id", order.ID)
	return nil
}

func (s *service) Submit(ctx context.Context, userID uint) (*Order, error) {
	order, err := s.repo.FindDraftByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		slog.Error("submit: find draft failed", "user_id", userID, "err", err)
		return nil, err
	}

	slog.Info("submitting order", "order_id", order.ID, "user_id", userID, "items", len(order.Items))

	if err := s.repo.UpdateStatus(ctx, order.ID, StatusSubmitted, nil); err != nil {
		slog.Error("submit: update status failed", "order_id", order.ID, "err", err)
		return nil, err
	}

	slog.Info("order submitted", "order_id", order.ID, "user_id", userID)

	for _, role := range []roles.Role{roles.RoleManager, roles.RoleAdmin} {
		staff, err := s.userRepo.FindAllByRole(ctx, role)
		if err != nil {
			slog.Error("submit: fetch staff failed", "role", role, "err", err)
			continue
		}
		for _, u := range staff {
			u := u
			s.tryNotify(ctx, order.ID, u.ID, func() error {
				return s.emailSender.SendOrderSubmitted(u.Email, order.ID)
			})
		}
	}

	order.Status = StatusSubmitted
	return order, nil
}

func (s *service) GetOrders(ctx context.Context, userID uint, role string, f Filter) ([]*Order, error) {
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if role == string(roles.RoleClient) {
		f.UserID = userID
	}
	return s.repo.FindFiltered(ctx, f)
}

func (s *service) GetOrder(ctx context.Context, userID uint, role string, orderID uint) (*Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ErrNotFound
	}
	if role == string(roles.RoleClient) && order.UserID != userID {
		slog.Warn("get order forbidden", "user_id", userID, "order_id", orderID)
		return nil, ErrForbidden
	}
	return order, nil
}

func (s *service) CancelOrder(ctx context.Context, userID uint, role string, orderID uint) error {
	if role != string(roles.RoleClient) {
		return ErrForbidden
	}

	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return ErrNotFound
	}
	if order.UserID != userID {
		slog.Warn("cancel order forbidden", "user_id", userID, "order_id", orderID, "owner_id", order.UserID)
		return ErrForbidden
	}
	if order.Status == StatusApproved || order.Status == StatusCancelled {
		slog.Warn("cancel order rejected", "order_id", orderID, "status", order.Status)
		return ErrInvalidStatus
	}

	if err := s.repo.UpdateStatus(ctx, orderID, StatusCancelled, nil); err != nil {
		slog.Error("cancel order failed", "order_id", orderID, "err", err)
		return err
	}
	slog.Info("order cancelled", "order_id", orderID, "user_id", userID)
	return nil
}

func (s *service) ApproveOrder(ctx context.Context, userID uint, role string, orderID uint, totalPrice float64) (*Order, error) {
	if role != string(roles.RoleManager) && role != string(roles.RoleAdmin) {
		return nil, ErrForbidden
	}

	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ErrNotFound
	}
	if order.Status != StatusSubmitted {
		slog.Warn("approve order rejected: wrong status", "order_id", orderID, "status", order.Status)
		return nil, ErrInvalidStatus
	}

	slog.Info("approving order", "order_id", orderID, "manager_id", userID, "total_price", totalPrice)

	if err := s.repo.UpdateStatus(ctx, orderID, StatusApproved, &totalPrice); err != nil {
		slog.Error("approve order failed", "order_id", orderID, "err", err)
		return nil, err
	}

	slog.Info("order approved", "order_id", orderID, "total_price", totalPrice)

	client, err := s.userRepo.FindByID(ctx, order.UserID)
	if err != nil {
		slog.Error("approve: fetch client failed", "client_id", order.UserID, "err", err)
	} else {
		s.tryNotify(ctx, order.ID, client.ID, func() error {
			return s.emailSender.SendOrderApproved(client.Email, order.ID, totalPrice)
		})
	}

	order.Status = StatusApproved
	order.TotalPrice = &totalPrice
	return order, nil
}

func (s *service) GetMessages(ctx context.Context, userID uint, role string, orderID uint) ([]Message, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ErrNotFound
	}
	if role == string(roles.RoleClient) && order.UserID != userID {
		slog.Warn("get messages forbidden", "user_id", userID, "order_id", orderID)
		return nil, ErrForbidden
	}
	return s.repo.FindMessages(ctx, orderID)
}

func (s *service) SendMessage(ctx context.Context, userID uint, role string, orderID uint, text string) (*Message, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, ErrNotFound
	}
	if role == string(roles.RoleClient) && order.UserID != userID {
		slog.Warn("send message forbidden", "user_id", userID, "order_id", orderID)
		return nil, ErrForbidden
	}
	if order.Status != StatusSubmitted && order.Status != StatusApproved {
		slog.Warn("send message rejected: chat not open", "order_id", orderID, "status", order.Status)
		return nil, ErrInvalidStatus
	}

	msg := &Message{OrderID: orderID, UserID: userID, Text: text}
	if err := s.repo.AddMessage(ctx, msg); err != nil {
		slog.Error("send message failed", "order_id", orderID, "user_id", userID, "err", err)
		return nil, err
	}
	slog.Info("message sent", "order_id", orderID, "user_id", userID, "msg_id", msg.ID)

	if role == string(roles.RoleClient) {
		for _, r := range []roles.Role{roles.RoleManager, roles.RoleAdmin} {
			staff, err := s.userRepo.FindAllByRole(ctx, r)
			if err != nil {
				slog.Error("send message: fetch staff failed", "role", r, "err", err)
				continue
			}
			for _, u := range staff {
				u := u
				s.tryNotify(ctx, order.ID, u.ID, func() error {
					return s.emailSender.SendNewMessage(u.Email, order.ID)
				})
			}
		}
	} else {
		client, err := s.userRepo.FindByID(ctx, order.UserID)
		if err != nil {
			slog.Error("send message: fetch client failed", "client_id", order.UserID, "err", err)
		} else {
			s.tryNotify(ctx, order.ID, client.ID, func() error {
				return s.emailSender.SendNewMessage(client.Email, order.ID)
			})
		}
	}

	return msg, nil
}

func (s *service) tryNotify(ctx context.Context, orderID, recipientID uint, send func() error) {
	key := fmt.Sprintf("notify:%d:%d", orderID, recipientID)
	_, err := s.rdb.SetArgs(ctx, key, 1, redis.SetArgs{
		Mode: "NX",
		TTL:  3 * 24 * time.Hour,
	}).Result()
	if errors.Is(err, redis.Nil) {
		slog.Debug("notification throttled", "order_id", orderID, "recipient_id", recipientID)
		return
	}
	if err != nil {
		slog.Error("notify throttle check failed", "err", err)
		return
	}
	if err := send(); err != nil {
		slog.Error("notification send failed", "order_id", orderID, "recipient_id", recipientID, "err", err)
		s.rdb.Del(ctx, key)
	}
}
