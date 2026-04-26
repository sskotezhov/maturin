package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
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
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidCode        = errors.New("invalid or expired code")
)

type EmailSender interface {
	SendVerificationCode(to, code string) error
	SendPasswordResetCode(to, code string) error
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	UserID       uint
	Email        string
	Role         user.Role
}

type Service interface {
	Register(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, email, code, password string) (*Tokens, error)
	ResendVerification(ctx context.Context, email string) error
	Login(ctx context.Context, email, password string) (*Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, code, newPassword string) error
}

type service struct {
	userRepo        user.Repository
	redis           *redis.Client
	emailSender     EmailSender
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(userRepo user.Repository, rdb *redis.Client, emailSender EmailSender, jwtSecret string, accessTTL, refreshTTL time.Duration) Service {
	return &service{
		userRepo:        userRepo,
		redis:           rdb,
		emailSender:     emailSender,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *service) Register(ctx context.Context, email string) error {
	slog.Info("user register started", "email", email)

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return ErrEmailAlreadyTaken
	}

	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	if err := s.redis.Set(ctx, emailVerifyKey(email), code, 15*time.Minute).Err(); err != nil {
		return err
	}

	if err := s.emailSender.SendVerificationCode(email, code); err != nil {
		return err
	}

	slog.Info("verification code sent", "email", email)
	return nil
}

func (s *service) ResendVerification(ctx context.Context, email string) error {
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return ErrEmailAlreadyTaken
	}

	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	if err := s.redis.Set(ctx, emailVerifyKey(email), code, 15*time.Minute).Err(); err != nil {
		return err
	}

	return s.emailSender.SendVerificationCode(email, code)
}

func (s *service) VerifyEmail(ctx context.Context, email, code, password string) (*Tokens, error) {
	stored, err := s.redis.Get(ctx, emailVerifyKey(email)).Result()
	if err != nil || stored != code {
		return nil, ErrInvalidCode
	}

	if err := util.ValidatePassword(password); err != nil {
		return nil, ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Email:         email,
		PasswordHash:  string(hash),
		Role:          user.RoleClient,
		EmailVerified: true,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	if err := s.redis.Del(ctx, emailVerifyKey(email)).Err(); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, u)
}

func (s *service) Login(ctx context.Context, email, password string) (*Tokens, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !u.EmailVerified {
		return nil, ErrEmailNotVerified
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

func (s *service) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil // не раскрываем существование email
	}

	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	if err := s.redis.Set(ctx, pwdResetKey(u.ID), code, 15*time.Minute).Err(); err != nil {
		return err
	}

	return s.emailSender.SendPasswordResetCode(email, code)
}

func (s *service) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrInvalidCredentials
	}

	stored, err := s.redis.Get(ctx, pwdResetKey(u.ID)).Result()
	if err != nil || stored != code {
		return ErrInvalidCode
	}

	if err := util.ValidatePassword(newPassword); err != nil {
		return ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}

	return s.redis.Del(ctx, pwdResetKey(u.ID)).Err()
}

func refreshKey(token string) string {
	return fmt.Sprintf("refresh:%s", token)
}

func emailVerifyKey(email string) string {
	return fmt.Sprintf("email_verify:%s", email)
}

func pwdResetKey(userID uint) string {
	return fmt.Sprintf("pwd_reset:%d", userID)
}

func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
