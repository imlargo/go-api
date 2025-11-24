package domain

import (
	"errors"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/validators"
)

// ValidateRegisterUser validates a RegisterUserRequest
func ValidateRegisterUser(request *dto.RegisterUserRequest) error {
	if request.Name == "" {
		return errors.New("name is required")
	}

	if request.Email == "" {
		return errors.New("email is required")
	}

	if err := validators.ValidateEmail(request.Email); err != nil {
		return err
	}

	if request.Password == "" {
		return errors.New("password is required")
	}

	if err := validators.ValidatePasswordWithContext(request.Password, request.Email, request.Name); err != nil {
		return err
	}

	if !request.UserType.IsValid() {
		return errors.New("invalid user type")
	}

	return nil
}

// ValidateChangePassword validates a ChangePasswordRequest
func ValidateChangePassword(request *dto.ChangePasswordRequest, userEmail, userName string) error {
	if request.OldPassword == "" {
		return errors.New("old password is required")
	}

	if request.NewPassword == "" {
		return errors.New("new password is required")
	}

	if request.NewPasswordConfirm == "" {
		return errors.New("password confirmation is required")
	}

	if request.NewPassword != request.NewPasswordConfirm {
		return errors.New("new password and confirmation do not match")
	}

	if err := validators.ValidatePasswordWithContext(request.NewPassword, userEmail, userName); err != nil {
		return err
	}

	return nil
}
