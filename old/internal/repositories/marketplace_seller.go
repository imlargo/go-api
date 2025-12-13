package repositories

import (
	"errors"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceSellerRepository interface {
	Create(seller *models.MarketplaceSeller) error
	GetByID(id uint) (*models.MarketplaceSeller, error)
	GetByUserID(userID uint) (*models.MarketplaceSeller, error)
	Update(seller *models.MarketplaceSeller) error
	Delete(id uint) error
	GetAll() ([]*models.MarketplaceSeller, error)
	IsUserSeller(userID uint) (bool, error)
}

type marketplaceSellerRepository struct {
	*Repository
}

func NewMarketplaceSellerRepository(
	r *Repository,
) MarketplaceSellerRepository {
	return &marketplaceSellerRepository{
		Repository: r,
	}
}

func (r *marketplaceSellerRepository) Create(seller *models.MarketplaceSeller) error {
	if err := r.db.Create(seller).Error; err != nil {
		return err
	}

	r.cache.Delete(r.cacheKeys.IsUserSeller(seller.UserID))

	return nil
}

func (r *marketplaceSellerRepository) GetByID(id uint) (*models.MarketplaceSeller, error) {
	var seller models.MarketplaceSeller
	if err := r.db.First(&seller, id).Error; err != nil {
		return nil, err
	}
	return &seller, nil
}

func (r *marketplaceSellerRepository) GetByUserID(userID uint) (*models.MarketplaceSeller, error) {
	var seller models.MarketplaceSeller
	if err := r.db.Where("user_id = ?", userID).First(&seller).Error; err != nil {
		return nil, err
	}
	return &seller, nil
}

func (r *marketplaceSellerRepository) Update(seller *models.MarketplaceSeller) error {
	return r.db.Model(seller).Clauses(clause.Returning{}).Updates(seller).Error
}

func (r *marketplaceSellerRepository) Delete(id uint) error {
	seller, err := r.GetByID(id)
	if err != nil {
		return err
	}

	if seller == nil {
		return errors.New("seller not found")
	}

	if err := r.db.Delete(&models.MarketplaceSeller{ID: id}).Error; err != nil {
		return err
	}

	r.cache.Delete(r.cacheKeys.IsUserSeller(seller.UserID))

	return nil
}

func (r *marketplaceSellerRepository) GetAll() ([]*models.MarketplaceSeller, error) {
	var sellers []*models.MarketplaceSeller
	if err := r.db.Find(&sellers).Error; err != nil {
		return nil, err
	}
	return sellers, nil
}

func (r *marketplaceSellerRepository) IsUserSeller(userID uint) (bool, error) {

	cacheKey := r.cacheKeys.IsUserSeller(userID)
	if cached, err := r.cache.GetBool(cacheKey); err == nil {
		return cached, nil
	}

	var count int64
	if err := r.db.Model(&models.MarketplaceSeller{}).Where(&models.MarketplaceSeller{UserID: userID}).Count(&count).Error; err != nil {
		return false, err
	}

	isSeller := count > 0

	if err := r.cache.Set(cacheKey, isSeller, 15*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return isSeller, nil
}
