package services

import (
	"errors"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *dto.RegisterUserRequest) (*models.User, error)
	DeleteUser(userID uint) error
	UpdateUser(userID uint, data *models.User) (*models.User, error)
	UpdateUserPreferences(userID uint, data *dto.UpdateUserPreferencesRequest) (*models.User, error)
	GetUserByID(userID uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}

type userServiceImpl struct {
	*Service
}

func NewUserService(container *Service) UserService {
	return &userServiceImpl{
		Service: container,
	}
}

func (s *userServiceImpl) CreateUser(data *dto.RegisterUserRequest) (*models.User, error) {

	data.Email = strings.ToLower(data.Email)

	// Validate user data
	user := &models.User{
		Name:      data.Name,
		Email:     data.Email,
		Password:  data.Password,
		Role:      data.UserType,
		CreatedBy: data.CreatedBy,
		TierLevel: enums.TierFree, // Users start on free tier
	}

	if data.UserType == "" {
		user.Role = enums.UserRoleAgency
	}

	if err := domain.ValidateRegisterUser(data); err != nil {
		return nil, err
	}

	existingUser, _ := s.store.Users.GetByEmail(user.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Validate and process referral code
	if data.ReferralCode != "" {
		referralCode, err := s.store.ReferralCode.GetByCode(data.ReferralCode)
		if err != nil {
			return nil, errors.New("invalid referral code")
		}

		if referralCode == nil {
			return nil, errors.New("invalid referral code")
		}

		// Check if code is active
		if referralCode.Status != enums.ReferralCodeStatusActive {
			return nil, errors.New("referral code is not active")
		}

		// Check if code has expired
		if referralCode.ExpiresAt != nil && referralCode.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("referral code has expired")
		}

		// Check usage limit
		if referralCode.UsageLimit != nil && referralCode.Registrations >= *referralCode.UsageLimit {
			return nil, errors.New("referral code has reached its usage limit")
		}

		user.ReferralCodeID = referralCode.ID
	}

	// Validate creator user
	if user.CreatedBy != 0 {
		creator, err := s.store.Users.GetByID(user.CreatedBy)
		if err != nil {
			return nil, errors.New("creator user not found")
		}

		if creator == nil {
			return nil, errors.New("creator user not found")
		}

		user.TierLevel = enums.TierFree // Created users always start on free tier
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	if err := s.store.Users.Create(user); err != nil {
		return nil, err
	}

	// Track referral registration
	if data.ReferralCode != "" {
		// Increment registrations counter
		go func() {
			err := s.store.ReferralCode.IncrementRegistrations(user.ReferralCodeID)
			if err != nil {
				s.logger.Errorf("failed to increment referral code registrations: %v", err)
			}
		}()
	}

	return user, nil
}

func (s *userServiceImpl) DeleteUser(userID uint) error {
	return nil
}

func (s *userServiceImpl) UpdateUserPreferences(userID uint, data *dto.UpdateUserPreferencesRequest) (*models.User, error) {
	// Get existing user
	existingUser, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Update preference fields
	existingUser.Type = data.Type
	existingUser.Industry = data.Industry
	existingUser.Goal = data.Goal
	existingUser.TeamSize = data.TeamSize

	// Update in database
	if err := s.store.Users.Update(existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *userServiceImpl) UpdateUser(userID uint, data *models.User) (*models.User, error) {
	// Get existing user
	existingUser, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Update only provided fields
	if data.Name != "" {
		existingUser.Name = data.Name
	}
	if data.Type != "" {
		existingUser.Type = data.Type
	}
	if data.Industry != "" {
		existingUser.Industry = data.Industry
	}
	if data.Goal != "" {
		existingUser.Goal = data.Goal
	}
	if data.TeamSize != "" {
		existingUser.TeamSize = data.TeamSize
	}

	// Update in database
	if err := s.store.Users.Update(existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *userServiceImpl) GetUserByID(userID uint) (*models.User, error) {
	return s.store.Users.GetByID(userID)
}

func (s *userServiceImpl) GetUserByEmail(email string) (*models.User, error) {
	return s.store.Users.GetByEmail(email)
}
