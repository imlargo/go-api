package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceRevisionRequestRepository interface {
	Create(revision *models.MarketplaceRevisionRequest) error
	GetByID(id uint) (*models.MarketplaceRevisionRequest, error)
	Update(revision *models.MarketplaceRevisionRequest) error
	Delete(id uint) error
	GetAllByOrder(orderID uint) ([]*models.MarketplaceRevisionRequest, error)
	GetPendingByOrder(orderID uint) (*models.MarketplaceRevisionRequest, error)
}

type marketplaceRevisionRequestRepository struct {
	*Repository
}

func NewMarketplaceRevisionRequestRepository(
	r *Repository,
) MarketplaceRevisionRequestRepository {
	return &marketplaceRevisionRequestRepository{
		Repository: r,
	}
}

func (r *marketplaceRevisionRequestRepository) Create(revision *models.MarketplaceRevisionRequest) error {
	return r.db.Create(revision).Error
}

func (r *marketplaceRevisionRequestRepository) GetByID(id uint) (*models.MarketplaceRevisionRequest, error) {
	var revision models.MarketplaceRevisionRequest
	if err := r.db.Preload("Order").Preload("Deliverable").First(&revision, id).Error; err != nil {
		return nil, err
	}
	return &revision, nil
}

func (r *marketplaceRevisionRequestRepository) Update(revision *models.MarketplaceRevisionRequest) error {
	return r.db.Model(revision).Clauses(clause.Returning{}).Updates(revision).Error
}

func (r *marketplaceRevisionRequestRepository) Delete(id uint) error {
	var revision models.MarketplaceRevisionRequest
	revision.ID = id
	return r.db.Delete(&revision).Error
}

func (r *marketplaceRevisionRequestRepository) GetAllByOrder(orderID uint) ([]*models.MarketplaceRevisionRequest, error) {
	var revisions []*models.MarketplaceRevisionRequest
	if err := r.db.Preload("Order").Preload("Deliverable").Where(&models.MarketplaceRevisionRequest{
		OrderID: orderID,
	}).Order("created_at DESC").Find(&revisions).Error; err != nil {
		return nil, err
	}
	return revisions, nil
}

func (r *marketplaceRevisionRequestRepository) GetPendingByOrder(orderID uint) (*models.MarketplaceRevisionRequest, error) {
	var revision models.MarketplaceRevisionRequest
	if err := r.db.Preload("Order").Preload("Deliverable").Where("order_id = ? AND status = ?", orderID, "pending").First(&revision).Error; err != nil {
		return nil, err
	}
	return &revision, nil
}
