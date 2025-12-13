package repositories

import (
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"

	"gorm.io/gorm/clause"
)

type MarketplaceCategoryRepository interface {
	Create(category *models.MarketplaceCategory) error
	GetByID(id uint) (*dto.MarketplaceCategoryResult, error)
	Update(category *models.MarketplaceCategory) error
	Delete(id uint) error
	GetAll() ([]*dto.MarketplaceCategoryResult, error)
}

type marketplaceCategoryRepository struct {
	*Repository
}

func NewMarketplaceCategoryRepository(
	r *Repository,
) MarketplaceCategoryRepository {
	return &marketplaceCategoryRepository{
		Repository: r,
	}
}

func (r *marketplaceCategoryRepository) Create(category *models.MarketplaceCategory) error {
	return r.db.Create(category).Error
}

func (r *marketplaceCategoryRepository) GetByID(id uint) (*dto.MarketplaceCategoryResult, error) {
	var result *dto.MarketplaceCategoryResult

	err := r.db.
		Model(&models.MarketplaceCategory{}).
		Select(`
			marketplace_categories.*,
			COALESCE(COUNT(DISTINCT marketplace_services.seller_id), 0) as sellers,
			COALESCE(COUNT(DISTINCT marketplace_orders.id), 0) as orders
		`).
		Joins("LEFT JOIN marketplace_services ON marketplace_categories.id = marketplace_services.category_id").
		Joins("LEFT JOIN marketplace_orders ON marketplace_categories.id = marketplace_orders.category_id").
		Where("marketplace_categories.id = ?", id).
		Group("marketplace_categories.id").
		First(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *marketplaceCategoryRepository) Update(category *models.MarketplaceCategory) error {
	return r.db.Model(category).Clauses(clause.Returning{}).Updates(category).Error
}

func (r *marketplaceCategoryRepository) Delete(id uint) error {
	var category models.MarketplaceCategory
	category.ID = id
	return r.db.Delete(&category).Error
}

func (r *marketplaceCategoryRepository) GetAll() ([]*dto.MarketplaceCategoryResult, error) {
	var results []*dto.MarketplaceCategoryResult

	err := r.db.
		Model(&models.MarketplaceCategory{}).
		Select(`
			marketplace_categories.*,
			COALESCE(COUNT(DISTINCT marketplace_services.seller_id), 0) as sellers,
			COALESCE(COUNT(DISTINCT marketplace_orders.id), 0) as orders
		`).
		Joins("LEFT JOIN marketplace_services ON marketplace_categories.id = marketplace_services.category_id").
		Joins("LEFT JOIN marketplace_orders ON marketplace_categories.id = marketplace_orders.category_id").
		Group("marketplace_categories.id").
		Find(&results).Error

	return results, err
}
