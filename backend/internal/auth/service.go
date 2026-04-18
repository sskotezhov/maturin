package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/util"
)

var (
	ErrEmailAlreadyTaken  = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrWeakPassword       = errors.New("password does not meet requirements")
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
	UserID       uint
	Email        string
	Role         user.Role
}

type Service interface {
	Register(ctx context.Context, email, password string) (*Tokens, error)
	Login(ctx context.Context, email, password string) (*Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

type service struct {
	userRepo        user.Repository
	redis           *redis.Client
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(userRepo user.Repository, rdb *redis.Client, jwtSecret string, accessTTL, refreshTTL time.Duration) Service {
	return &service{
		userRepo:        userRepo,
		redis:           rdb,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *service) Register(ctx context.Context, email, password string) (*Tokens, error) {
	slog.Info("user register started", "email", email)
	if err := util.ValidatePassword(password); err != nil {
		return nil, ErrWeakPassword
	}

	slog.Info("check for unique email", "email", email)
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		slog.Info("email already taken", "email", email)
		return nil, ErrEmailAlreadyTaken
	}
	slog.Info("email is unique", "email", email)

	slog.Info("hashing password")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	slog.Info("password hashed")

	u := &user.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         user.RoleClient,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	tokens, err := s.issueTokens(ctx, u)
	if err != nil {
		return nil, err
	}

	slog.Info("user register finished", "email", email)
	return tokens, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*Tokens, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, u)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	key := refreshKey(refreshToken)

	userID, err := s.redis.Get(ctx, key).Uint64()
	if err != nil {
		return nil, ErrInvalidToken
	}

	u, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, u)
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	return s.redis.Del(ctx, refreshKey(refreshToken)).Err()
}

// issueTokens generates a new access + refresh token pair for the given user.
func (s *service) issueTokens(ctx context.Context, u *user.User) (*Tokens, error) {
	accessToken, err := s.newAccessToken(u)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.newRefreshToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       u.ID,
		Email:        u.Email,
		Role:         u.Role,
	}, nil
}

func (s *service) newAccessToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"role": string(u.Role),
		"exp":  time.Now().Add(s.accessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *service) newRefreshToken(ctx context.Context, userID uint) (string, error) {
	token, err := util.GenerateToken()
	if err != nil {
		return "", err
	}

	if err := s.redis.Set(ctx, refreshKey(token), userID, s.refreshTokenTTL).Err(); err != nil {
		return "", err
	}

	return token, nil
}

func refreshKey(token string) string {
	return fmt.Sprintf("refresh:%s", token)
}
