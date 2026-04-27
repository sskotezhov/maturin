package staff

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

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
	ErrNotFound       = errors.New("not found")
	ErrInvalidRole    = errors.New("invalid role")
	ErrSelfRoleChange = errors.New("cannot change own role")
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
}

type service struct {
	userRepo   user.Repository
	orderRepo  order.Repository
	productSvc productService
}

func NewService(userRepo user.Repository, orderRepo order.Repository, productSvc productService) Service {
	return &service{
		userRepo:   userRepo,
		orderRepo:  orderRepo,
		productSvc: productSvc,
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
