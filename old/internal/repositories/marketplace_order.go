package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type MarketplaceOrderRepository interface {
	Create(order *models.MarketplaceOrder) error
	GetByID(id uint) (*models.MarketplaceOrder, error)
	Update(order *models.MarketplaceOrder) error
	Delete(id uint) error
	GetAll() ([]*models.MarketplaceOrder, error)
	GetAllByClient(clientId uint) ([]*models.MarketplaceOrder, error)
	GetAllByBuyer(userId uint) ([]*models.MarketplaceOrder, error)
	GetAllBySeller(sellerID uint) ([]*models.MarketplaceOrder, error)
	GetOrdersForAutoCompletion() ([]*models.MarketplaceOrder, error)
}

type marketplaceOrderRepository struct {
	*Repository
}

func NewMarketplaceOrderRepository(
	r *Repository,
) MarketplaceOrderRepository {
	return &marketplaceOrderRepository{
		Repository: r,
	}
}

func (r *marketplaceOrderRepository) Create(order *models.MarketplaceOrder) error {
	return r.db.Create(order).Error
}

func (r *marketplaceOrderRepository) GetByID(id uint) (*models.MarketplaceOrder, error) {
	var order models.MarketplaceOrder
	if err := r.db.Preload("Category").Preload("Service").Preload("Service.Seller").Preload("Service.User").Preload("Buyer").Preload("Client").Preload("ServicePackage").Preload("Payment").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *marketplaceOrderRepository) Update(order *models.MarketplaceOrder) error {
	return r.db.Model(order).Clauses(clause.Returning{}).Updates(order).Error
}

func (r *marketplaceOrderRepository) Delete(id uint) error {
	var order models.MarketplaceOrder
	order.ID = id
	return r.db.Delete(&order).Error
}

func (r *marketplaceOrderRepository) GetAll() ([]*models.MarketplaceOrder, error) {
	var orders []*models.MarketplaceOrder
	if err := r.db.Preload("Category").Preload("Service").Preload("Service.Seller").Preload("Service.User").Preload("Buyer").Preload("Client").Preload("ServicePackage").Preload("Payment").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *marketplaceOrderRepository) GetAllByClient(clientID uint) ([]*models.MarketplaceOrder, error) {
	var orders []*models.MarketplaceOrder
	if err := r.db.Preload("Category").Preload("Service").Preload("Service.Seller").Preload("Service.User").Preload("Buyer").Preload("Client").Preload("ServicePackage").Preload("Payment").Where(&models.MarketplaceOrder{
		ClientID: clientID,
	}).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *marketplaceOrderRepository) GetAllByBuyer(userID uint) ([]*models.MarketplaceOrder, error) {
	var orders []*models.MarketplaceOrder
	if err := r.db.Preload("Category").Preload("Service").Preload("Service.Seller").Preload("Service.User").Preload("Buyer").Preload("Client").Preload("ServicePackage").Preload("Payment").Where(&models.MarketplaceOrder{
		BuyerID: userID,
	}).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *marketplaceOrderRepository) GetAllBySeller(sellerID uint) ([]*models.MarketplaceOrder, error) {
	var orders []*models.MarketplaceOrder
	if err := r.db.Preload("Category").Preload("Service").Preload("Service.Seller").Preload("Service.User").Preload("Buyer").Preload("Client").Preload("ServicePackage").Preload("Payment").Joins("Service").
		Where("\"Service\".user_id = ?", sellerID).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *marketplaceOrderRepository) GetOrdersForAutoCompletion() ([]*models.MarketplaceOrder, error) {
	var orders []*models.MarketplaceOrder
	if err := r.db.Preload("Service").Preload("Service.Seller").Preload("Service.User").Preload("Buyer").
		Where("status = ? AND auto_completion_date <= ?", "delivered", time.Now()).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
