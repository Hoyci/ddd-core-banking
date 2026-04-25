package valueobjects

import (
	"ddd-core-banking/internal/onboarding/domain"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))

	if normalized == "" {
		return Email{}, domain.ErrEmailIsRequired
	}
	if len(normalized) > 254 {
		return Email{}, domain.ErrExceedLength
	}
	if !emailRegex.MatchString(normalized) {
		return Email{}, domain.ErrInvalidEmail
	}

	return Email{value: normalized}, nil
}

func (e Email) Value() string  { return e.value }
func (e Email) String() string { return e.value }
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}
