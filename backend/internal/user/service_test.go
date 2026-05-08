package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sskotezhov/maturin/internal/user"
)

func newTestUserService(t *testing.T) (user.Service, *mockUserRepository) {
	t.Helper()
	repo := &mockUserRepository{}
	svc := user.NewService(repo)
	return svc, repo
}

// GetProfile

func TestGetProfile_UserNotFound(t *testing.T) {
	svc, repo := newTestUserService(t)
	repo.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("not found"))

	_, err := svc.GetProfile(context.Background(), 1)
	assert.ErrorIs(t, err, user.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T) {
	svc, repo := newTestUserService(t)
	u := &user.User{ID: 1, Email: "user@example.com", FirstName: "Иван"}
	repo.On("FindByID", mock.Anything, uint(1)).Return(u, nil)

	result, err := svc.GetProfile(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", result.Email)
	assert.Equal(t, "Иван", result.FirstName)
	repo.AssertExpectations(t)
}

// UpdateProfile

func TestUpdateProfile_ForbiddenField_Email(t *testing.T) {
	svc, _ := newTestUserService(t)

	_, err := svc.UpdateProfile(context.Background(), 1, user.UpdateInput{
		Mask:   []string{"email"},
		Values: map[string]string{"email": "new@example.com"},
	})
	assert.ErrorIs(t, err, user.ErrForbiddenField)
}

func TestUpdateProfile_ForbiddenField_Role(t *testing.T) {
	svc, _ := newTestUserService(t)

	_, err := svc.UpdateProfile(context.Background(), 1, user.UpdateInput{
		Mask:   []string{"role"},
		Values: map[string]string{"role": "admin"},
	})
	assert.ErrorIs(t, err, user.ErrForbiddenField)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	svc, repo := newTestUserService(t)
	repo.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("not found"))

	_, err := svc.UpdateProfile(context.Background(), 1, user.UpdateInput{
		Mask:   []string{"first_name"},
		Values: map[string]string{"first_name": "Сергей"},
	})
	assert.ErrorIs(t, err, user.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateProfile_AllowedFields_Success(t *testing.T) {
	svc, repo := newTestUserService(t)
	u := &user.User{ID: 1, Email: "user@example.com", FirstName: "Иван"}
	repo.On("FindByID", mock.Anything, uint(1)).Return(u, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	result, err := svc.UpdateProfile(context.Background(), 1, user.UpdateInput{
		Mask:   []string{"first_name", "phone"},
		Values: map[string]string{"first_name": "Сергей", "phone": "+79991234567"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Сергей", result.FirstName)
	assert.Equal(t, "+79991234567", result.Phone)
	repo.AssertExpectations(t)
}

func TestUpdateProfile_EmptyMask_NoChanges(t *testing.T) {
	svc, repo := newTestUserService(t)
	u := &user.User{ID: 1, Email: "user@example.com", FirstName: "Иван"}
	repo.On("FindByID", mock.Anything, uint(1)).Return(u, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	result, err := svc.UpdateProfile(context.Background(), 1, user.UpdateInput{
		Mask:   []string{},
		Values: map[string]string{},
	})
	require.NoError(t, err)
	assert.Equal(t, "Иван", result.FirstName, "name should be unchanged")
	repo.AssertExpectations(t)
}

func TestUpdateProfile_AllAllowedFields(t *testing.T) {
	svc, repo := newTestUserService(t)
	u := &user.User{ID: 1}
	repo.On("FindByID", mock.Anything, uint(1)).Return(u, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	result, err := svc.UpdateProfile(context.Background(), 1, user.UpdateInput{
		Mask: []string{"last_name", "first_name", "middle_name", "phone", "telegram", "company_name", "inn"},
		Values: map[string]string{
			"last_name":    "Иванов",
			"first_name":   "Иван",
			"middle_name":  "Иванович",
			"phone":        "+79991234567",
			"telegram":     "@ivan",
			"company_name": "ООО Ромашка",
			"inn":          "1234567890",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Иванов", result.LastName)
	assert.Equal(t, "Иван", result.FirstName)
	assert.Equal(t, "Иванович", result.MiddleName)
	assert.Equal(t, "+79991234567", result.Phone)
	assert.Equal(t, "@ivan", result.Telegram)
	assert.Equal(t, "ООО Ромашка", result.CompanyName)
	assert.Equal(t, "1234567890", result.INN)
	repo.AssertExpectations(t)
}
