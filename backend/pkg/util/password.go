package util

import (
	"errors"
	"unicode"
)

var (
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrPasswordNoDigit        = errors.New("password must contain at least one digit")
	ErrPasswordNonLatinLetter = errors.New("password must contain only latin letters")
)

// ValidatePassword enforces: length >= 8, at least one digit, no non-latin letters.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	hasDigit := false
	for _, c := range password {
		if unicode.IsLetter(c) && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return ErrPasswordNonLatinLetter
		}
		if unicode.IsDigit(c) {
			hasDigit = true
		}
	}

	if !hasDigit {
		return ErrPasswordNoDigit
	}

	return nil
}
