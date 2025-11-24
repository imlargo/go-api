package validators

import (
	"errors"
	"strings"
)

// ValidatePassword validates a password string against security requirements
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
	insecurePasswords := []string{"12345678", "password", "admin123"}
	for _, insecurePassword := range insecurePasswords {
		if cleanedPassword == insecurePassword {
			return errors.New("password is too weak")
		}
	}

	return nil
}

// ValidatePasswordWithContext validates a password with additional context (email, name)
func ValidatePasswordWithContext(password, email, name string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	cleanedPassword := strings.TrimSpace(password)

	if cleanedPassword == email {
		return errors.New("password cannot be the same as email")
	}

	if cleanedPassword == name {
		return errors.New("password cannot be the same as name")
	}

	return nil
}
