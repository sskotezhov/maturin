package inquiry_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sskotezhov/maturin/internal/inquiry"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

func newTestInquiryService(t *testing.T) (inquiry.Service, *mockInquiryRepo, *mockInquiryUserRepo, *mockInquiryEmailSender) {
	t.Helper()
	repo := &mockInquiryRepo{}
	userRepo := &mockInquiryUserRepo{}
	emailSender := &mockInquiryEmailSender{}
	svc := inquiry.NewService(repo, userRepo, emailSender)
	return svc, repo, userRepo, emailSender
}

func validInput() inquiry.SubmitInput {
	return inquiry.SubmitInput{
		Name:    "Иван Иванов",
		Phone:   "+7 (999) 123-45-67",
		Comment: "Хочу узнать о продукции",
		Consent: true,
	}
}

func TestSubmit_ConsentFalse(t *testing.T) {
	svc, _, _, _ := newTestInquiryService(t)
	input := validInput()
	input.Consent = false

	_, err := svc.Submit(context.Background(), input)
	assert.ErrorIs(t, err, inquiry.ErrConsentRequired)
}

func TestSubmit_NameTooShort(t *testing.T) {
	svc, _, _, _ := newTestInquiryService(t)
	input := validInput()
	input.Name = "А"

	_, err := svc.Submit(context.Background(), input)
	assert.ErrorIs(t, err, inquiry.ErrInvalidName)
}

func TestSubmit_NameTooLong(t *testing.T) {
	svc, _, _, _ := newTestInquiryService(t)
	input := validInput()
	input.Name = strings.Repeat("А", 101)

	_, err := svc.Submit(context.Background(), input)
	assert.ErrorIs(t, err, inquiry.ErrInvalidName)
}

func TestSubmit_PhoneTooFewDigits(t *testing.T) {
	svc, _, _, _ := newTestInquiryService(t)
	input := validInput()
	input.Phone = "123"

	_, err := svc.Submit(context.Background(), input)
	assert.ErrorIs(t, err, inquiry.ErrInvalidPhone)
}

func TestSubmit_CommentTooLong(t *testing.T) {
	svc, _, _, _ := newTestInquiryService(t)
	input := validInput()
	input.Comment = strings.Repeat("а", 2001)

	_, err := svc.Submit(context.Background(), input)
	assert.ErrorIs(t, err, inquiry.ErrCommentTooLong)
}

func TestSubmit_Success_SavesAndNotifiesStaff(t *testing.T) {
	svc, repo, userRepo, emailSender := newTestInquiryService(t)
	input := validInput()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*inquiry.Inquiry")).Return(nil)
	manager := &user.User{ID: 1, Email: "manager@example.com", Role: roles.RoleManager}
	userRepo.On("FindAllByRole", mock.Anything, roles.RoleManager).Return([]*user.User{manager}, nil)
	userRepo.On("FindAllByRole", mock.Anything, roles.RoleAdmin).Return([]*user.User{}, nil)
	emailSender.On("SendInquirySubmitted", "manager@example.com",
		mock.AnythingOfType("uint"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(nil)

	result, err := svc.Submit(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, inquiry.StatusNew, result.Status)
	assert.True(t, result.ConsentAccepted)
	repo.AssertExpectations(t)
	emailSender.AssertExpectations(t)
}

func TestSubmit_StaffDeduplication_AdminAndManagerSameUser(t *testing.T) {
	// If the same user appears in both manager and admin role queries,
	// the inquiry service deduplicates by user ID.
	svc, repo, userRepo, emailSender := newTestInquiryService(t)
	input := validInput()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*inquiry.Inquiry")).Return(nil)
	// User 1 is manager
	manager := &user.User{ID: 1, Email: "staff@example.com", Role: roles.RoleManager}
	// User 1 also appears as admin (e.g., has both roles in different queries)
	admin := &user.User{ID: 1, Email: "staff@example.com", Role: roles.RoleAdmin}
	userRepo.On("FindAllByRole", mock.Anything, roles.RoleManager).Return([]*user.User{manager}, nil)
	userRepo.On("FindAllByRole", mock.Anything, roles.RoleAdmin).Return([]*user.User{admin}, nil)
	// Should send only once due to dedup
	emailSender.On("SendInquirySubmitted", "staff@example.com",
		mock.AnythingOfType("uint"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(nil).Once()

	_, err := svc.Submit(context.Background(), input)
	require.NoError(t, err)
	emailSender.AssertExpectations(t)
}

func TestSubmit_NameOnlySpaces_InvalidName(t *testing.T) {
	svc, _, _, _ := newTestInquiryService(t)
	input := validInput()
	input.Name = "   "

	_, err := svc.Submit(context.Background(), input)
	assert.ErrorIs(t, err, inquiry.ErrInvalidName)
}
