package services

import (
	"errors"

	"github.com/imlargo/go-api-template/internal/dto"
	"github.com/imlargo/go-api-template/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *dto.RegisterUser) (*models.User, error)
	DeleteUser(userID uint) error
	UpdateUser(userID uint, data *models.User) (*models.User, error)
	GetUserByID(userID uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}

type userService struct {
	*Service
}

func NewUserService(service *Service) UserService {
	return &userService{
		Service: service,
	}
}

func (s *userService) CreateUser(data *dto.RegisterUser) (*models.User, error) {
	// Validate user data
	user := &models.User{
		Name:     data.Name,
		Email:    data.Email,
		Password: data.Password,
	}

	if err := user.ValidateUserCreation(); err != nil {
		return nil, err
	}

	existingUser, _ := s.store.Users.GetByEmail(user.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	if err := user.ValidatePassword(); err != nil {
		return nil, err
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

	return user, nil
}

func (s *userService) DeleteUser(userID uint) error {
	return nil
}

func (s *userService) UpdateUser(userID uint, data *models.User) (*models.User, error) {
	return nil, nil
}

func (s *userService) GetUserByID(userID uint) (*models.User, error) {
	return nil, nil
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return nil, nil
}
