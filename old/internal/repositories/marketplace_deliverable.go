package repositories

import (
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceDeliverableRepository interface {
	Create(deliverable *models.MarketplaceDeliverable) error
	GetByID(id uint) (*models.MarketplaceDeliverable, error)
	Update(deliverable *models.MarketplaceDeliverable) error
	Delete(id uint) error
	GetAllByOrder(orderID uint) ([]*models.MarketplaceDeliverable, error)
	GetPendingByOrder(orderID uint) (*models.MarketplaceDeliverable, error)
}

type marketplaceDeliverableRepository struct {
	*Repository
}

func NewMarketplaceDeliverableRepository(
	r *Repository,
) MarketplaceDeliverableRepository {
	return &marketplaceDeliverableRepository{
		Repository: r,
	}
}

func (r *marketplaceDeliverableRepository) Create(deliverable *models.MarketplaceDeliverable) error {
	return r.db.Create(deliverable).Error
}

func (r *marketplaceDeliverableRepository) GetByID(id uint) (*models.MarketplaceDeliverable, error) {
	var deliverable models.MarketplaceDeliverable
	if err := r.db.Preload("Order").Preload("Seller").First(&deliverable, id).Error; err != nil {
		return nil, err
	}
	return &deliverable, nil
}

func (r *marketplaceDeliverableRepository) Update(deliverable *models.MarketplaceDeliverable) error {
	return r.db.Model(deliverable).Clauses(clause.Returning{}).Updates(deliverable).Error
}

func (r *marketplaceDeliverableRepository) Delete(id uint) error {
	var deliverable models.MarketplaceDeliverable
	deliverable.ID = id
	return r.db.Delete(&deliverable).Error
}

func (r *marketplaceDeliverableRepository) GetAllByOrder(orderID uint) ([]*models.MarketplaceDeliverable, error) {
	var deliverables []*models.MarketplaceDeliverable
	if err := r.db.Preload("Order").Preload("Seller").Where(&models.MarketplaceDeliverable{
		OrderID: orderID,
	}).Order("created_at DESC").Find(&deliverables).Error; err != nil {
		return nil, err
	}
	return deliverables, nil
}

func (r *marketplaceDeliverableRepository) GetPendingByOrder(orderID uint) (*models.MarketplaceDeliverable, error) {
	var deliverable models.MarketplaceDeliverable
	if err := r.db.Preload("Order").Preload("Seller").Where(&models.MarketplaceDeliverable{
		OrderID: orderID,
		Status:  enums.DeliverableStatusPending,
	}).First(&deliverable).Error; err != nil {
		return nil, err
	}
	return &deliverable, nil
}
