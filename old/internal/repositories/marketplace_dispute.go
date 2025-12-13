package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceDisputeRepository interface {
	Create(dispute *models.MarketplaceDispute) error
	GetByID(id uint) (*models.MarketplaceDispute, error)
	Update(dispute *models.MarketplaceDispute) error
	Delete(id uint) error
	GetAll() ([]*models.MarketplaceDispute, error)
	GetByOrder(orderID uint) (*models.MarketplaceDispute, error)
	GetAllByStatus(status string) ([]*models.MarketplaceDispute, error)
	GetAllBySeller(sellerID uint) ([]*models.MarketplaceDispute, error)
	GetAllByBuyer(buyerID uint) ([]*models.MarketplaceDispute, error)
}

type marketplaceDisputeRepository struct {
	*Repository
}

func NewMarketplaceDisputeRepository(
	r *Repository,
) MarketplaceDisputeRepository {
	return &marketplaceDisputeRepository{
		Repository: r,
	}
}

func (r *marketplaceDisputeRepository) Create(dispute *models.MarketplaceDispute) error {
	return r.db.Create(dispute).Error
}

func (r *marketplaceDisputeRepository) GetByID(id uint) (*models.MarketplaceDispute, error) {
	var dispute models.MarketplaceDispute
	if err := r.db.Preload("Order").Preload("Order.Service").Preload("Opener").Preload("Resolver").First(&dispute, id).Error; err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *marketplaceDisputeRepository) Update(dispute *models.MarketplaceDispute) error {
	return r.db.Model(dispute).Clauses(clause.Returning{}).Updates(dispute).Error
}

func (r *marketplaceDisputeRepository) Delete(id uint) error {
	var dispute models.MarketplaceDispute
	dispute.ID = id
	return r.db.Delete(&dispute).Error
}

func (r *marketplaceDisputeRepository) GetAll() ([]*models.MarketplaceDispute, error) {
	var disputes []*models.MarketplaceDispute
	if err := r.db.Preload("Order").Preload("Order.Service").Preload("Opener").Preload("Resolver").Order("created_at DESC").Find(&disputes).Error; err != nil {
		return nil, err
	}
	return disputes, nil
}

func (r *marketplaceDisputeRepository) GetByOrder(orderID uint) (*models.MarketplaceDispute, error) {
	var dispute models.MarketplaceDispute
	if err := r.db.Preload("Order").Preload("Order.Service").Preload("Opener").Preload("Resolver").Where(&models.MarketplaceDispute{
		OrderID: orderID,
	}).First(&dispute).Error; err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *marketplaceDisputeRepository) GetAllByStatus(status string) ([]*models.MarketplaceDispute, error) {
	var disputes []*models.MarketplaceDispute
	if err := r.db.Preload("Order").Preload("Order.Service").Preload("Opener").Preload("Resolver").Where("status = ?", status).Order("created_at DESC").Find(&disputes).Error; err != nil {
		return nil, err
	}
	return disputes, nil
}

func (r *marketplaceDisputeRepository) GetAllBySeller(sellerID uint) ([]*models.MarketplaceDispute, error) {
	var disputes []*models.MarketplaceDispute
	if err := r.db.Preload("Order").Preload("Order.Service").Preload("Opener").Preload("Resolver").
		Joins("JOIN marketplace_orders ON marketplace_orders.id = marketplace_disputes.order_id").
		Joins("JOIN marketplace_services ON marketplace_services.id = marketplace_orders.service_id").
		Where("marketplace_services.user_id = ?", sellerID).
		Order("marketplace_disputes.created_at DESC").
		Find(&disputes).Error; err != nil {
		return nil, err
	}
	return disputes, nil
}

func (r *marketplaceDisputeRepository) GetAllByBuyer(buyerID uint) ([]*models.MarketplaceDispute, error) {
	var disputes []*models.MarketplaceDispute
	if err := r.db.Preload("Order").Preload("Order.Service").Preload("Opener").Preload("Resolver").
		Joins("JOIN marketplace_orders ON marketplace_orders.id = marketplace_disputes.order_id").
		Where("marketplace_orders.buyer_id = ?", buyerID).
		Order("marketplace_disputes.created_at DESC").
		Find(&disputes).Error; err != nil {
		return nil, err
	}
	return disputes, nil
}
