package inquiry

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/roles"
)

const (
	maxNameLen    = 100
	maxPhoneLen   = 40
	maxCommentLen = 2000
	maxSourceLen  = 100
	maxPageURLLen = 500
)

var (
	ErrConsentRequired = errors.New("consent required")
	ErrInvalidName     = errors.New("invalid name")
	ErrInvalidPhone    = errors.New("invalid phone")
	ErrCommentTooLong  = errors.New("comment too long")
)

type EmailSender interface {
	SendInquirySubmitted(to string, inquiryID uint, name, phone, comment string) error
}

type SubmitInput struct {
	Name    string
	Phone   string
	Comment string
	Consent bool
	Source  string
	PageURL string
}

type Service interface {
	Submit(ctx context.Context, input SubmitInput) (*Inquiry, error)
}

type service struct {
	repo        Repository
	userRepo    user.Repository
	emailSender EmailSender
}

func NewService(repo Repository, userRepo user.Repository, emailSender EmailSender) Service {
	return &service{
		repo:        repo,
		userRepo:    userRepo,
		emailSender: emailSender,
	}
}

func (s *service) Submit(ctx context.Context, input SubmitInput) (*Inquiry, error) {
	if !input.Consent {
		return nil, ErrConsentRequired
	}

	name := compactSpaces(input.Name)
	if len([]rune(name)) < 2 || len([]rune(name)) > maxNameLen {
		return nil, ErrInvalidName
	}

	phone := strings.TrimSpace(input.Phone)
	if len([]rune(phone)) > maxPhoneLen {
		return nil, ErrInvalidPhone
	}
	phoneDigits := digitsOnly(phone)
	if len(phoneDigits) < 10 || len(phoneDigits) > 15 {
		return nil, ErrInvalidPhone
	}

	comment := strings.TrimSpace(input.Comment)
	if len([]rune(comment)) > maxCommentLen {
		return nil, ErrCommentTooLong
	}

	item := &Inquiry{
		Name:            name,
		Phone:           phone,
		PhoneDigits:     phoneDigits,
		Comment:         comment,
		Source:          limitString(strings.TrimSpace(input.Source), maxSourceLen),
		PageURL:         limitString(strings.TrimSpace(input.PageURL), maxPageURLLen),
		ConsentAccepted: true,
		Status:          StatusNew,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		slog.Error("inquiry: create failed", "err", err)
		return nil, err
	}

	slog.Info("inquiry: submitted", "inquiry_id", item.ID)
	s.notifyStaff(ctx, item)

	return item, nil
}

func (s *service) notifyStaff(ctx context.Context, item *Inquiry) {
	seen := make(map[uint]struct{})
	for _, role := range []roles.Role{roles.RoleManager, roles.RoleAdmin} {
		staff, err := s.userRepo.FindAllByRole(ctx, role)
		if err != nil {
			slog.Error("inquiry: fetch staff failed", "role", role, "err", err)
			continue
		}
		for _, u := range staff {
			if _, ok := seen[u.ID]; ok {
				continue
			}
			seen[u.ID] = struct{}{}
			if err := s.emailSender.SendInquirySubmitted(u.Email, item.ID, item.Name, item.Phone, item.Comment); err != nil {
				slog.Error("inquiry: notify failed", "inquiry_id", item.ID, "recipient_id", u.ID, "err", err)
			}
		}
	}
}

func compactSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func limitString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
