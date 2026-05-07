package slot

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	slotDuration = 30 * time.Minute
	maxNameLen   = 100
	minNameLen   = 2
	minPhoneLen  = 10
	maxPhoneLen  = 15
)

var (
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrSlotBooked        = errors.New("cannot delete booked slot")
	ErrInvalidRange      = errors.New("end must be after start + 30 min")
	ErrInvalidName       = errors.New("invalid name")
	ErrInvalidPhone      = errors.New("invalid phone")
)

type RangeInput struct {
	Start     time.Time
	End       time.Time
	ManagerID uint
}

type BookInput struct {
	Name  string
	Phone string
}

type Service interface {
	CreateRanges(ctx context.Context, ranges []RangeInput) (int, error)
	ListAvailable(ctx context.Context, f AvailableFilter) ([]*CallSlot, error)
	Book(ctx context.Context, slotID uint, input BookInput) (*CallBooking, error)
	StaffListSlots(ctx context.Context, f Filter) ([]*CallSlot, int, error)
	StaffDeleteSlot(ctx context.Context, id uint) error
	StaffListBookings(ctx context.Context, f BookingFilter) ([]*CallBooking, int, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateRanges(ctx context.Context, ranges []RangeInput) (int, error) {
	var all []CallSlot
	for _, r := range ranges {
		slots := sliceIntoSlots(r.Start, r.End, r.ManagerID)
		if len(slots) == 0 {
			return 0, ErrInvalidRange
		}
		all = append(all, slots...)
	}
	if len(all) == 0 {
		return 0, ErrInvalidRange
	}
	if err := s.repo.CreateBatch(ctx, all); err != nil {
		return 0, err
	}
	return len(all), nil
}

func (s *service) ListAvailable(ctx context.Context, f AvailableFilter) ([]*CallSlot, error) {
	return s.repo.FindAvailable(ctx, f)
}

func (s *service) Book(ctx context.Context, slotID uint, input BookInput) (*CallBooking, error) {
	name := strings.Join(strings.Fields(strings.TrimSpace(input.Name)), " ")
	if len([]rune(name)) < minNameLen || len([]rune(name)) > maxNameLen {
		return nil, ErrInvalidName
	}

	phone := strings.TrimSpace(input.Phone)
	digits := digitsOnly(phone)
	if len(digits) < minPhoneLen || len(digits) > maxPhoneLen {
		return nil, ErrInvalidPhone
	}

	if _, err := s.repo.FindByID(ctx, slotID); err != nil {
		return nil, ErrSlotNotFound
	}

	booking := &CallBooking{
		Name:        name,
		Phone:       phone,
		PhoneDigits: digits,
	}
	if err := s.repo.Book(ctx, slotID, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *service) StaffListSlots(ctx context.Context, f Filter) ([]*CallSlot, int, error) {
	return s.repo.FindFiltered(ctx, f)
}

func (s *service) StaffDeleteSlot(ctx context.Context, id uint) error {
	return s.repo.DeleteByID(ctx, id)
}

func (s *service) StaffListBookings(ctx context.Context, f BookingFilter) ([]*CallBooking, int, error) {
	return s.repo.FindBookings(ctx, f)
}

func sliceIntoSlots(start, end time.Time, managerID uint) []CallSlot {
	var slots []CallSlot
	for s := start; !s.Add(slotDuration).After(end); s = s.Add(slotDuration) {
		slots = append(slots, CallSlot{
			ManagerID: managerID,
			StartAt:   s,
			EndAt:     s.Add(slotDuration),
		})
	}
	return slots
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
