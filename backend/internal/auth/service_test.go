package auth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/sskotezhov/maturin/internal/auth"
	"github.com/sskotezhov/maturin/internal/user"
)

func newTestAuthService(t *testing.T) (auth.Service, *mockUserRepo, *mockAuthEmailSender, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	userRepo := &mockUserRepo{}
	emailSender := &mockAuthEmailSender{}
	svc := auth.NewService(userRepo, rdb, emailSender, "test-secret", 15*time.Minute, 7*24*time.Hour)
	return svc, userRepo, emailSender, mr
}

// Register

func TestRegister_EmailAlreadyTaken(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	existing := &user.User{ID: 1, Email: "taken@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "taken@example.com").Return(existing, nil)

	err := svc.Register(context.Background(), "taken@example.com")
	assert.ErrorIs(t, err, auth.ErrEmailAlreadyTaken)
	userRepo.AssertExpectations(t)
}

func TestRegister_NewEmail_SetsRedisKeyAndSendsEmail(t *testing.T) {
	svc, userRepo, emailSender, mr := newTestAuthService(t)
	userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
	emailSender.On("SendVerificationCode", "new@example.com", mock.AnythingOfType("string")).Return(nil)

	err := svc.Register(context.Background(), "new@example.com")
	require.NoError(t, err)

	key := "email_verify:new@example.com"
	val, err2 := mr.Get(key)
	require.NoError(t, err2)
	assert.NotEmpty(t, val, "Redis key should be set after Register")
	userRepo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

// VerifyEmail

func TestVerifyEmail_CodeNotInRedis(t *testing.T) {
	svc, _, _, _ := newTestAuthService(t)

	_, err := svc.VerifyEmail(context.Background(), "test@example.com", "123456", "ValidPass1!")
	assert.ErrorIs(t, err, auth.ErrInvalidCode)
}

func TestVerifyEmail_WrongCode(t *testing.T) {
	svc, _, _, mr := newTestAuthService(t)
	require.NoError(t, mr.Set("email_verify:test@example.com", "654321"))

	_, err := svc.VerifyEmail(context.Background(), "test@example.com", "123456", "ValidPass1!")
	assert.ErrorIs(t, err, auth.ErrInvalidCode)
}

func TestVerifyEmail_WeakPassword(t *testing.T) {
	svc, _, _, mr := newTestAuthService(t)
	require.NoError(t, mr.Set("email_verify:test@example.com", "123456"))

	_, err := svc.VerifyEmail(context.Background(), "test@example.com", "123456", "weak")
	assert.ErrorIs(t, err, auth.ErrWeakPassword)
}

func TestVerifyEmail_Success(t *testing.T) {
	svc, userRepo, _, mr := newTestAuthService(t)
	require.NoError(t, mr.Set("email_verify:test@example.com", "123456"))
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	tokens, err := svc.VerifyEmail(context.Background(), "test@example.com", "123456", "ValidPass1!")
	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// Key should be deleted after successful verification
	_, getErr := mr.Get("email_verify:test@example.com")
	assert.Error(t, getErr, "Redis key should be deleted after verification")
	userRepo.AssertExpectations(t)
}

// Login

func TestLogin_EmailNotFound(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	userRepo.On("FindByEmail", mock.Anything, "nobody@example.com").Return(nil, errors.New("not found"))

	_, err := svc.Login(context.Background(), "nobody@example.com", "anypassword")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1!"), bcrypt.MinCost)
	u := &user.User{ID: 1, Email: "user@example.com", PasswordHash: string(hash), EmailVerified: true}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)

	_, err := svc.Login(context.Background(), "user@example.com", "WrongPass1!")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestLogin_EmailNotVerified(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("ValidPass1!"), bcrypt.MinCost)
	u := &user.User{ID: 1, Email: "user@example.com", PasswordHash: string(hash), EmailVerified: false}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)

	_, err := svc.Login(context.Background(), "user@example.com", "ValidPass1!")
	assert.ErrorIs(t, err, auth.ErrEmailNotVerified)
	userRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("ValidPass1!"), bcrypt.MinCost)
	u := &user.User{ID: 1, Email: "user@example.com", PasswordHash: string(hash), EmailVerified: true, Role: user.RoleClient}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)

	tokens, err := svc.Login(context.Background(), "user@example.com", "ValidPass1!")
	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, uint(1), tokens.UserID)
	userRepo.AssertExpectations(t)
}

// Refresh

func TestRefresh_InvalidToken(t *testing.T) {
	svc, _, _, _ := newTestAuthService(t)

	_, err := svc.Refresh(context.Background(), "nonexistent-token")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestRefresh_Success(t *testing.T) {
	svc, userRepo, _, mr := newTestAuthService(t)
	require.NoError(t, mr.Set("refresh:validtoken", "42"))
	u := &user.User{ID: 42, Email: "user@example.com", Role: user.RoleClient}
	userRepo.On("FindByID", mock.Anything, uint(42)).Return(u, nil)

	tokens, err := svc.Refresh(context.Background(), "validtoken")
	require.NoError(t, err)
	require.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)

	// Old token should be deleted
	_, getErr := mr.Get("refresh:validtoken")
	assert.Error(t, getErr, "old refresh token should be deleted")
	userRepo.AssertExpectations(t)
}

// Logout

func TestLogout_DeletesRefreshToken(t *testing.T) {
	svc, _, _, mr := newTestAuthService(t)
	require.NoError(t, mr.Set("refresh:mytoken", "1"))

	err := svc.Logout(context.Background(), "mytoken")
	require.NoError(t, err)
	_, getErr := mr.Get("refresh:mytoken")
	assert.Error(t, getErr, "refresh token should be deleted after logout")
}

// ForgotPassword

func TestForgotPassword_EmailNotFound_ReturnsNil(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	userRepo.On("FindByEmail", mock.Anything, "ghost@example.com").Return(nil, errors.New("not found"))

	err := svc.ForgotPassword(context.Background(), "ghost@example.com")
	assert.NoError(t, err, "should not reveal that email doesn't exist")
	userRepo.AssertExpectations(t)
}

func TestForgotPassword_EmailFound_SetsKeyAndSendsEmail(t *testing.T) {
	svc, userRepo, emailSender, mr := newTestAuthService(t)
	u := &user.User{ID: 5, Email: "user@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)
	emailSender.On("SendPasswordResetCode", "user@example.com", mock.AnythingOfType("string")).Return(nil)

	err := svc.ForgotPassword(context.Background(), "user@example.com")
	require.NoError(t, err)

	key := fmt.Sprintf("pwd_reset:%d", u.ID)
	val, getErr := mr.Get(key)
	require.NoError(t, getErr)
	assert.NotEmpty(t, val)
	userRepo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

// ResetPassword

func TestResetPassword_EmailNotFound(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	userRepo.On("FindByEmail", mock.Anything, "nobody@example.com").Return(nil, errors.New("not found"))

	err := svc.ResetPassword(context.Background(), "nobody@example.com", "123456", "NewPass1!")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_WrongCode(t *testing.T) {
	svc, userRepo, _, _ := newTestAuthService(t)
	u := &user.User{ID: 7, Email: "user@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)
	// No Redis key set → wrong code

	err := svc.ResetPassword(context.Background(), "user@example.com", "123456", "NewPass1!")
	assert.ErrorIs(t, err, auth.ErrInvalidCode)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_WeakPassword(t *testing.T) {
	svc, userRepo, _, mr := newTestAuthService(t)
	u := &user.User{ID: 7, Email: "user@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)
	require.NoError(t, mr.Set("pwd_reset:7", "123456"))

	err := svc.ResetPassword(context.Background(), "user@example.com", "123456", "weak")
	assert.ErrorIs(t, err, auth.ErrWeakPassword)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_Success(t *testing.T) {
	svc, userRepo, _, mr := newTestAuthService(t)
	u := &user.User{ID: 7, Email: "user@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	require.NoError(t, mr.Set("pwd_reset:7", "123456"))

	err := svc.ResetPassword(context.Background(), "user@example.com", "123456", "NewPass1!")
	require.NoError(t, err)
	_, getErr := mr.Get("pwd_reset:7")
	assert.Error(t, getErr, "Redis key should be deleted after password reset")
	userRepo.AssertExpectations(t)
}
