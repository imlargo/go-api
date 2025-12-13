package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceServiceResultRepository interface {
	Create(result *models.MarketplaceServiceResult) error
	GetByID(id uint) (*models.MarketplaceServiceResult, error)
	Update(result *models.MarketplaceServiceResult) error
	Delete(id uint) error
	GetAll() ([]*models.MarketplaceServiceResult, error)
	GetAllByService(serviceId uint) ([]*models.MarketplaceServiceResult, error)
}

type marketplaceServiceResultRepository struct {
	*Repository
}

func NewMarketplaceServiceResultRepository(
	r *Repository,
) MarketplaceServiceResultRepository {
	return &marketplaceServiceResultRepository{
		Repository: r,
	}
}

func (r *marketplaceServiceResultRepository) Create(result *models.MarketplaceServiceResult) error {
	return r.db.Create(result).Error
}

func (r *marketplaceServiceResultRepository) GetByID(id uint) (*models.MarketplaceServiceResult, error) {
	var result models.MarketplaceServiceResult
	if err := r.db.Preload("File").First(&result, id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *marketplaceServiceResultRepository) Update(result *models.MarketplaceServiceResult) error {
	return r.db.Model(result).Clauses(clause.Returning{}).Updates(result).Error
}

func (r *marketplaceServiceResultRepository) Delete(id uint) error {
	var result models.MarketplaceServiceResult
	result.ID = id
	return r.db.Delete(&result).Error
}

func (r *marketplaceServiceResultRepository) GetAll() ([]*models.MarketplaceServiceResult, error) {
	var results []*models.MarketplaceServiceResult
	if err := r.db.Preload("File").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *marketplaceServiceResultRepository) GetAllByService(serviceId uint) ([]*models.MarketplaceServiceResult, error) {
	var results []*models.MarketplaceServiceResult
	if err := r.db.Preload("File").Where("service_id = ?", serviceId).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
