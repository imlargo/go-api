package repositories

import (
	"github.com/google/uuid"
	"github.com/imlargo/go-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	GetAll() ([]*models.User, error)
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) (*models.User, error)
	Delete(id uuid.UUID) error
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

func (s *UserRepositoryImpl) Create(user *models.User) (*models.User, error) {

	result := s.db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (service *UserRepositoryImpl) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := service.db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (service *UserRepositoryImpl) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := service.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserRepositoryImpl) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := service.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserRepositoryImpl) Update(user *models.User) (*models.User, error) {
	return nil, nil
}
func (service *UserRepositoryImpl) Delete(id uuid.UUID) error {
	return nil
}
