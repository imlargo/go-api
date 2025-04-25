package services

import (
	"github.com/google/uuid"
	"github.com/imlargo/go-api/internal/models"
	"github.com/imlargo/go-api/internal/store"
)

type UserService interface {
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Create(user *models.User) (*models.User, error)
	Update(user *models.User) (*models.User, error)
	Delete(id uuid.UUID) error
}

type UserServiceImpl struct {
	storage *store.Storage
}

func NewUserService(storage *store.Storage) UserService {
	return &UserServiceImpl{
		storage: storage,
	}
}

func (service *UserServiceImpl) Create(user *models.User) (*models.User, error) {

	user, err := service.storage.Users.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (service *UserServiceImpl) GetAll() ([]*models.User, error) {
	users, err := service.storage.Users.GetAll()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (service *UserServiceImpl) GetByID(id uuid.UUID) (*models.User, error) {
	user, err := service.storage.Users.GetByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (service *UserServiceImpl) GetByEmail(email string) (*models.User, error) {
	user, err := service.storage.Users.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (service *UserServiceImpl) Update(user *models.User) (*models.User, error) {
	return nil, nil
}
func (service *UserServiceImpl) Delete(id uuid.UUID) error {
	return nil
}
