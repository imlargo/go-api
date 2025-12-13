package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

type ReferralDiscountRepository interface {
	Create(discount *models.ReferralDiscount) error
	GetByUserID(userID uint) (*models.ReferralDiscount, error)
	HasUserUsedDiscount(userID uint) (bool, error)
	GetByID(id uint) (*models.ReferralDiscount, error)
	Update(discount *models.ReferralDiscount) error
}

type referralDiscountRepository struct {
	*Repository
}

func NewReferralDiscountRepository(r *Repository) ReferralDiscountRepository {
	return &referralDiscountRepository{Repository: r}
}

func (repo *referralDiscountRepository) Create(discount *models.ReferralDiscount) error {
	return repo.db.Create(discount).Error
}

func (repo *referralDiscountRepository) GetByUserID(userID uint) (*models.ReferralDiscount, error) {
	var discount models.ReferralDiscount
	err := repo.db.Where("user_id = ?", userID).First(&discount).Error
	if err != nil {
		return nil, err
	}
	return &discount, nil
}

func (repo *referralDiscountRepository) HasUserUsedDiscount(userID uint) (bool, error) {
	var discount models.ReferralDiscount
	err := repo.db.Where("user_id = ?", userID).Limit(1).First(&discount).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *referralDiscountRepository) GetByID(id uint) (*models.ReferralDiscount, error) {
	var discount models.ReferralDiscount
	err := repo.db.First(&discount, id).Error
	if err != nil {
		return nil, err
	}
	return &discount, nil
}

func (repo *referralDiscountRepository) Update(discount *models.ReferralDiscount) error {
	return repo.db.Save(discount).Error
}
