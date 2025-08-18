package models

import (
	"errors"
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`
}

func (user *User) ValidateUserCreation() error {
	if user.Name == "" {
		return errors.New("Name is required")
	}

	if user.Email == "" {
		return errors.New("Email is required")
	}

	if user.Password == "" {
		return errors.New("Password is required")
	}

	return nil
}

func (user *User) ValidatePassword() error {
	if user.Password == "" {
		return errors.New("Password is required")
	}

	if len(user.Password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}

	if len(user.Password) > 30 {
		return errors.New("Password must be less than 100 characters long")
	}

	if user.Password == user.Email {
		return errors.New("Password cannot be the same as email")
	}

	if user.Password == user.Name {
		return errors.New("Password cannot be the same as name")
	}

	// Check insecure passwords
	insecurePasswords := []string{"12345678", "password", "admin123"}
	for _, insecurePassword := range insecurePasswords {
		if user.Password == insecurePassword {
			return errors.New("Password is too weak")
		}
	}

	return nil
}
