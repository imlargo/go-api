package services

import (
	"errors"

	requestsdto "github.com/imlargo/go-api-template/internal/dto/requests"
	"github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *requestsdto.RegisterUserRequest) (*models.User, error)
	DeleteUser(userID uint) error
	UpdateUser(userID uint, data *models.User) (*models.User, error)
	GetUserByID(userID uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}

type userServiceImpl struct {
	store *store.Store
}

func NewUserService(store *store.Store) UserService {
	return &userServiceImpl{
		store,
	}
}

func (s *userServiceImpl) CreateUser(data *requestsdto.RegisterUserRequest) (*models.User, error) {
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

func (s *userServiceImpl) DeleteUser(userID uint) error {
	return nil
}

func (s *userServiceImpl) UpdateUser(userID uint, data *models.User) (*models.User, error) {
	return nil, nil
}

func (s *userServiceImpl) GetUserByID(userID uint) (*models.User, error) {
	return nil, nil
}

func (s *userServiceImpl) GetUserByEmail(email string) (*models.User, error) {
	return nil, nil
}
