package validators

import (
	"errors"
	"net/mail"
	"strings"
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) == 0 {
		return errors.New("email cannot be empty")
	}

	if len(email) > 254 {
		return errors.New("email cannot exceed 254 characters")
	}

	if !strings.Contains(email, "@") {
		return errors.New("invalid email format")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}
	return nil
}
