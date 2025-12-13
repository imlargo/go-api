package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceServicePackageRepository interface {
	Create(servicePackage *models.MarketplaceServicePackage) error
	GetByID(id uint) (*models.MarketplaceServicePackage, error)
	Update(servicePackage *models.MarketplaceServicePackage) error
	Delete(id uint) error
	GetAllByService(serviceId uint) ([]*models.MarketplaceServicePackage, error)
}
type marketplaceServicePackageRepository struct {
	*Repository
}

func NewMarketplaceServicePackageRepository(
	r *Repository,
) MarketplaceServicePackageRepository {
	return &marketplaceServicePackageRepository{
		Repository: r,
	}
}

func (r *marketplaceServicePackageRepository) Create(servicePackage *models.MarketplaceServicePackage) error {
	return r.db.Create(servicePackage).Error
}

func (r *marketplaceServicePackageRepository) GetByID(id uint) (*models.MarketplaceServicePackage, error) {
	var servicePackage models.MarketplaceServicePackage
	if err := r.db.First(&servicePackage, id).Error; err != nil {
		return nil, err
	}
	return &servicePackage, nil
}

func (r *marketplaceServicePackageRepository) Update(servicePackage *models.MarketplaceServicePackage) error {
	return r.db.Model(servicePackage).Clauses(clause.Returning{}).Updates(servicePackage).Error
}

func (r *marketplaceServicePackageRepository) Delete(id uint) error {
	var servicePackage models.MarketplaceServicePackage
	servicePackage.ID = id
	return r.db.Delete(&servicePackage).Error
}

func (r *marketplaceServicePackageRepository) GetAllByService(serviceId uint) ([]*models.MarketplaceServicePackage, error) {
	var servicePackages []*models.MarketplaceServicePackage
	if err := r.db.Where(&models.MarketplaceServicePackage{ServiceID: serviceId}).Find(&servicePackages).Error; err != nil {
		return nil, err
	}
	return servicePackages, nil
}
