package validators

import (
	"errors"
	"strings"
)

func ValidatePassword(password string) error {

	cleanedPassword := strings.TrimSpace(password)

	if cleanedPassword == "" {
		return errors.New("password cannot be empty")
	}

	if len(cleanedPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(cleanedPassword) > 100 {
		return errors.New("password cannot exceed 100 characters")
	}

	// Check insecure passwords
	insecurePasswords := []string{"12345678", "password"}
	for _, insecurePassword := range insecurePasswords {
		if cleanedPassword == insecurePassword {
			return errors.New("password is too weak")
		}
	}

	return nil
}
