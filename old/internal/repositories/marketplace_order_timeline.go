package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
)

type MarketplaceOrderTimelineRepository interface {
	Create(timeline *models.MarketplaceOrderTimeline) error
	GetByID(id uint) (*models.MarketplaceOrderTimeline, error)
	GetAllByOrder(orderID uint) ([]*models.MarketplaceOrderTimeline, error)
	Delete(id uint) error
}

type marketplaceOrderTimelineRepository struct {
	*Repository
}

func NewMarketplaceOrderTimelineRepository(
	r *Repository,
) MarketplaceOrderTimelineRepository {
	return &marketplaceOrderTimelineRepository{
		Repository: r,
	}
}

func (r *marketplaceOrderTimelineRepository) Create(timeline *models.MarketplaceOrderTimeline) error {
	return r.db.Create(timeline).Error
}

func (r *marketplaceOrderTimelineRepository) GetByID(id uint) (*models.MarketplaceOrderTimeline, error) {
	var timeline models.MarketplaceOrderTimeline
	if err := r.db.Preload("Order").Preload("Actor").First(&timeline, id).Error; err != nil {
		return nil, err
	}
	return &timeline, nil
}

func (r *marketplaceOrderTimelineRepository) GetAllByOrder(orderID uint) ([]*models.MarketplaceOrderTimeline, error) {
	var timelines []*models.MarketplaceOrderTimeline
	if err := r.db.Preload("Order").Preload("Actor").Where(&models.MarketplaceOrderTimeline{
		OrderID: orderID,
	}).Order("created_at ASC").Find(&timelines).Error; err != nil {
		return nil, err
	}
	return timelines, nil
}

func (r *marketplaceOrderTimelineRepository) Delete(id uint) error {
	var timeline models.MarketplaceOrderTimeline
	timeline.ID = id
	return r.db.Delete(&timeline).Error
}
