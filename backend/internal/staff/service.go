package staff

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/sskotezhov/maturin/internal/inquiry"
	"github.com/sskotezhov/maturin/internal/order"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

const (
	staleThreshold     = 72 * time.Hour
	clientOrdersLimit  = 20
	cacheRefreshBudget = 2 * time.Minute
)

var (
	ErrNotFound             = errors.New("not found")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidInquiryStatus = errors.New("invalid inquiry status")
	ErrSelfRoleChange       = errors.New("cannot change own role")
)

type ClientFilter struct {
	Q             string
	Role          string
	EmailVerified *bool
	Page          int
	Limit         int
}

type ClientDetails struct {
	Client      *user.User
	Orders      []*order.Order
	OrdersCount int
}

type Dashboard struct {
	OrdersByStatus      map[string]int
	StaleSubmittedCount int
}

type CacheStats struct {
	RefreshedAt     time.Time
	ProductsCount   int
	CategoriesCount int
}

type productService interface {
	RefreshCache(ctx context.Context) (int, int, error)
}

type Service interface {
	ListClients(ctx context.Context, f ClientFilter) ([]*user.User, int, error)
	GetClient(ctx context.Context, id uint) (*ClientDetails, error)
	Dashboard(ctx context.Context) (*Dashboard, error)
	ChangeRole(ctx context.Context, actorID, targetID uint, newRole roles.Role) error
	RefreshCatalog(ctx context.Context) (*CacheStats, error)
	ListInquiries(ctx context.Context, f inquiry.Filter) ([]*inquiry.Inquiry, int, error)
	GetInquiry(ctx context.Context, id uint) (*inquiry.Inquiry, error)
	ChangeInquiryStatus(ctx context.Context, id uint, status inquiry.Status) error
}

type service struct {
	userRepo    user.Repository
	orderRepo   order.Repository
	productSvc  productService
	inquiryRepo inquiry.Repository
}

func NewService(
	userRepo user.Repository,
	orderRepo order.Repository,
	productSvc productService,
	inquiryRepo inquiry.Repository,
) Service {
	return &service{
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		productSvc:  productSvc,
		inquiryRepo: inquiryRepo,
	}
}

func (s *service) ListClients(ctx context.Context, f ClientFilter) ([]*user.User, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page < 1 {
		f.Page = 1
	}
	users, total, err := s.userRepo.FindFiltered(ctx, user.Filter{
		Q:             f.Q,
		Role:          f.Role,
		EmailVerified: f.EmailVerified,
		Page:          f.Page,
		Limit:         f.Limit,
	})
	if err != nil {
		slog.Error("staff: list clients failed", "err", err)
		return nil, 0, err
	}
	return users, total, nil
}

func (s *service) GetClient(ctx context.Context, id uint) (*ClientDetails, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		slog.Error("staff: find user failed", "user_id", id, "err", err)
		return nil, err
	}

	orders, err := s.orderRepo.FindByUserID(ctx, id, clientOrdersLimit)
	if err != nil {
		slog.Error("staff: list user orders failed", "user_id", id, "err", err)
		return nil, err
	}
	s.enrichOrderResponseStatuses(ctx, orders)

	count, err := s.orderRepo.CountByUserID(ctx, id)
	if err != nil {
		slog.Error("staff: count user orders failed", "user_id", id, "err", err)
		return nil, err
	}

	return &ClientDetails{
		Client:      u,
		Orders:      orders,
		OrdersCount: count,
	}, nil
}

func (s *service) Dashboard(ctx context.Context) (*Dashboard, error) {
	byStatus, err := s.orderRepo.CountByStatus(ctx)
	if err != nil {
		slog.Error("staff: count by status failed", "err", err)
		return nil, err
	}
	stale, err := s.orderRepo.CountStaleSubmitted(ctx, staleThreshold)
	if err != nil {
		slog.Error("staff: count stale submitted failed", "err", err)
		return nil, err
	}

	out := make(map[string]int, len(byStatus))
	for k, v := range byStatus {
		out[string(k)] = v
	}

	return &Dashboard{
		OrdersByStatus:      out,
		StaleSubmittedCount: stale,
	}, nil
}

func (s *service) ChangeRole(ctx context.Context, actorID, targetID uint, newRole roles.Role) error {
	if !roles.Valid(string(newRole)) {
		return ErrInvalidRole
	}
	if actorID == targetID {
		return ErrSelfRoleChange
	}

	u, err := s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		slog.Error("staff: find target user failed", "target_id", targetID, "err", err)
		return err
	}

	u.Role = newRole
	if err := s.userRepo.Update(ctx, u); err != nil {
		slog.Error("staff: update role failed", "target_id", targetID, "err", err)
		return err
	}
	slog.Info("staff: role changed", "actor_id", actorID, "target_id", targetID, "new_role", newRole)
	return nil
}

func (s *service) enrichOrderResponseStatuses(ctx context.Context, orders []*order.Order) {
	for _, o := range orders {
		o.ResponseStatus = order.ResponseNone
		if o.Status != order.StatusSubmitted {
			continue
		}

		msg, err := s.orderRepo.FindLastMessage(ctx, o.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				o.ResponseStatus = order.ResponseWaitingManager
				continue
			}
			slog.Error("staff: find last order message failed", "order_id", o.ID, "err", err)
			continue
		}

		author, err := s.userRepo.FindByID(ctx, msg.UserID)
		if err != nil {
			slog.Error("staff: find message author failed", "order_id", o.ID, "user_id", msg.UserID, "err", err)
			continue
		}
		if author.Role == roles.RoleClient {
			o.ResponseStatus = order.ResponseWaitingManager
		} else {
			o.ResponseStatus = order.ResponseWaitingClient
		}
	}
}

func (s *service) RefreshCatalog(ctx context.Context) (*CacheStats, error) {
	ctx, cancel := context.WithTimeout(ctx, cacheRefreshBudget)
	defer cancel()

	products, cats, err := s.productSvc.RefreshCache(ctx)
	if err != nil {
		slog.Error("staff: refresh catalog failed", "err", err)
		return nil, err
	}
	slog.Info("staff: catalog refreshed", "products", products, "categories", cats)

	return &CacheStats{
		RefreshedAt:     time.Now(),
		ProductsCount:   products,
		CategoriesCount: cats,
	}, nil
}

func (s *service) ListInquiries(ctx context.Context, f inquiry.Filter) ([]*inquiry.Inquiry, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Status != "" && !inquiry.ValidStatus(f.Status) {
		return nil, 0, ErrInvalidInquiryStatus
	}

	items, total, err := s.inquiryRepo.FindFiltered(ctx, f)
	if err != nil {
		slog.Error("staff: list inquiries failed", "err", err)
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) GetInquiry(ctx context.Context, id uint) (*inquiry.Inquiry, error) {
	item, err := s.inquiryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		slog.Error("staff: find inquiry failed", "inquiry_id", id, "err", err)
		return nil, err
	}
	return item, nil
}

func (s *service) ChangeInquiryStatus(ctx context.Context, id uint, status inquiry.Status) error {
	if !inquiry.ValidStatus(string(status)) {
		return ErrInvalidInquiryStatus
	}
	if err := s.inquiryRepo.UpdateStatus(ctx, id, status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		slog.Error("staff: update inquiry status failed", "inquiry_id", id, "status", status, "err", err)
		return err
	}
	slog.Info("staff: inquiry status changed", "inquiry_id", id, "status", status)
	return nil
}
