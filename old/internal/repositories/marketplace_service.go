package repositories

import (
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MarketplaceServiceRepository interface {
	Create(service *models.MarketplaceService) error
	GetByID(id uint) (*dto.MarketplaceServiceDto, error)
	GetRawByID(id uint) (*models.MarketplaceService, error)
	Update(service *models.MarketplaceService) error
	UpdateImageID(serviceID uint, imageID uint) error
	Delete(id uint) error
	GetAllByCategory(categoryId uint) ([]*dto.MarketplaceServiceDto, error)
	GetAllBySearch(search string) ([]*dto.MarketplaceServiceDto, error)
	GetAllBySeller(sellerID uint) ([]*dto.MarketplaceServiceDto, error)
	DecreaseAvailableSpots(serviceID uint) error
}

type marketplaceServiceRepository struct {
	*Repository
}

func NewMarketplaceServiceRepository(
	r *Repository,
) MarketplaceServiceRepository {
	return &marketplaceServiceRepository{
		Repository: r,
	}
}

func (r *marketplaceServiceRepository) Create(service *models.MarketplaceService) error {
	return r.db.Create(service).Error
}

func (r *marketplaceServiceRepository) GetByID(id uint) (*dto.MarketplaceServiceDto, error) {
	var service dto.MarketplaceServiceDto

	err := r.db.
		Model(&models.MarketplaceService{}).
		Select(`
			marketplace_services.*,
			COALESCE(MIN(marketplace_service_packages.price), 0) as min_price,
			COALESCE(MAX(marketplace_service_packages.price), 0) as max_price,
			COALESCE(COUNT(DISTINCT marketplace_orders.id), 0) as orders_count
		`).
		Joins("LEFT JOIN marketplace_service_packages ON marketplace_services.id = marketplace_service_packages.service_id").
		Joins("LEFT JOIN marketplace_orders ON marketplace_services.id = marketplace_orders.service_id").
		Preload("Image").
		Preload("Seller").
		Where("marketplace_services.id = ?", id).
		Group("marketplace_services.id").
		First(&service).Error

	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *marketplaceServiceRepository) GetRawByID(id uint) (*models.MarketplaceService, error) {
	var service models.MarketplaceService

	err := r.db.
		Preload("Image").
		Preload("Seller").
		First(&service, id).Error

	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *marketplaceServiceRepository) Update(service *models.MarketplaceService) error {
	return r.db.Model(service).Clauses(clause.Returning{}).Updates(service).Error
}

// UpdateImageID updates only the image_id field for a marketplace service.
// This method performs a direct column update using a SQL UPDATE statement.
// Returns gorm.ErrRecordNotFound if the service with the given ID doesn't exist.
func (r *marketplaceServiceRepository) UpdateImageID(serviceID uint, imageID uint) error {
	result := r.db.Model(&models.MarketplaceService{}).
		Where("id = ?", serviceID).
		Update("image_id", imageID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *marketplaceServiceRepository) Delete(id uint) error {
	var service models.MarketplaceService
	service.ID = id
	return r.db.Delete(&service).Error
}

func (r *marketplaceServiceRepository) GetAllByCategory(categoryID uint) ([]*dto.MarketplaceServiceDto, error) {
	var results []*dto.MarketplaceServiceDto

	err := r.db.
		Model(&models.MarketplaceService{}).
		Select(`
			marketplace_services.*,
			COALESCE(MIN(marketplace_service_packages.price), 0) as min_price,
			COALESCE(MAX(marketplace_service_packages.price), 0) as max_price,
			COALESCE(COUNT(DISTINCT marketplace_orders.id), 0) as orders_count
		`).
		Joins("LEFT JOIN marketplace_service_packages ON marketplace_services.id = marketplace_service_packages.service_id").
		Joins("LEFT JOIN marketplace_orders ON marketplace_services.id = marketplace_orders.service_id").
		Preload("Image").
		Preload("Seller").
		Where(&models.MarketplaceService{CategoryID: categoryID}).
		Group("marketplace_services.id").
		Find(&results).Error

	return results, err
}

func (r *marketplaceServiceRepository) GetAllBySearch(search string) ([]*dto.MarketplaceServiceDto, error) {
	var results []*dto.MarketplaceServiceDto

	searchPattern := "%" + search + "%"

	err := r.db.
		Model(&models.MarketplaceService{}).
		Select(`
			marketplace_services.*,
			COALESCE(MIN(marketplace_service_packages.price), 0) as min_price,
			COALESCE(MAX(marketplace_service_packages.price), 0) as max_price,
			COALESCE(COUNT(DISTINCT marketplace_orders.id), 0) as orders_count
		`).
		Joins("LEFT JOIN marketplace_service_packages ON marketplace_services.id = marketplace_service_packages.service_id").
		Joins("LEFT JOIN marketplace_orders ON marketplace_services.id = marketplace_orders.service_id").
		Preload("Image").
		Preload("Seller").
		Where("title ILIKE ?", searchPattern).
		Group("marketplace_services.id").
		Find(&results).Error

	return results, err
}

func (r *marketplaceServiceRepository) GetAllBySeller(sellerID uint) ([]*dto.MarketplaceServiceDto, error) {
	var results []*dto.MarketplaceServiceDto

	err := r.db.
		Model(&models.MarketplaceService{}).
		Select(`
			marketplace_services.*,
			COALESCE(MIN(marketplace_service_packages.price), 0) as min_price,
			COALESCE(MAX(marketplace_service_packages.price), 0) as max_price,
			COALESCE(COUNT(DISTINCT marketplace_orders.id), 0) as orders_count
		`).
		Joins("LEFT JOIN marketplace_service_packages ON marketplace_services.id = marketplace_service_packages.service_id").
		Joins("LEFT JOIN marketplace_orders ON marketplace_services.id = marketplace_orders.service_id").
		Preload("Image").
		Preload("Seller").
		Where(&models.MarketplaceService{SellerID: sellerID}).
		Group("marketplace_services.id").
		Find(&results).Error

	return results, err
}

func (r *marketplaceServiceRepository) DecreaseAvailableSpots(serviceID uint) error {
	return r.db.Model(&models.MarketplaceService{}).
		Where(&models.MarketplaceService{ID: serviceID}).
		UpdateColumn("available_spots", gorm.Expr("available_spots - 1")).Error
}
