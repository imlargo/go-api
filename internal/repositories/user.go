package repositories

import (
	"log"
	"time"

	"github.com/imlargo/go-api/internal/models"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Create(user *models.User) error
	GetAll() ([]*models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepository struct {
	*Repository
}

func NewUserRepository(
	r *Repository,
) UserRepository {
	return &userRepository{
		Repository: r,
	}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {

	var user models.User

	cacheKey := r.cacheKeys.UserByID(id)
	if err := r.cache.GetJSON(cacheKey, &user); err == nil {
		return &user, nil
	}

	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	if err := r.cache.Set(cacheKey, &user, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {

	if err := r.db.Model(user).Clauses(clause.Returning{}).Updates(user).Error; err != nil {
		return err
	}

	r.invalidateCache(user.ID)

	return nil
}

func (r *userRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		return err
	}

	r.invalidateCache(id)

	return nil
}

func (r *userRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) invalidateCache(userID uint) {
	cacheKey := r.cacheKeys.UserByID(userID)
	if err := r.cache.Delete(cacheKey); err != nil {
		log.Println("Cache delete failed:", err.Error())
	}
}
