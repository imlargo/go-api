package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
)

type SubscriptionRepository interface {
	Create(subscription *models.Subscription) error
	GetByID(id uint) (*models.Subscription, error)
	GetByStripeID(stripeSubID string) (*models.Subscription, error)
	GetUserSubscriptions(userID uint) ([]*models.Subscription, error)
	GetUsersSubscriptions(userIDs []uint) ([]*models.Subscription, error)
	GetActiveUserSubscriptions(userID uint) ([]*models.Subscription, error)
	GetUserTierSubscription(userID uint) (*models.Subscription, error)
	GetUserAddonSubscriptions(userID uint) ([]*models.Subscription, error)
	Update(subscription *models.Subscription) error
	Delete(id uint) error
	GetAll(status string, subscriptionType string, limit, offset int) ([]*models.Subscription, int64, error)
}

type subscriptionRepository struct {
	*Repository
}

func NewSubscriptionRepository(r *Repository) SubscriptionRepository {
	return &subscriptionRepository{
		Repository: r,
	}
}

// Create creates a new subscription record
func (r *subscriptionRepository) Create(subscription *models.Subscription) error {
	return r.db.Create(subscription).Error
}

// GetByID retrieves a subscription by ID
func (r *subscriptionRepository) GetByID(id uint) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Preload("User").First(&subscription, id).Error
	return &subscription, err
}

// GetByStripeID retrieves a subscription by Stripe subscription ID
func (r *subscriptionRepository) GetByStripeID(stripeSubID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Preload("User").Where("stripe_subscription_id = ?", stripeSubID).First(&subscription).Error
	return &subscription, err
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (r *subscriptionRepository) GetUserSubscriptions(userID uint) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&subscriptions).Error
	return subscriptions, err
}

// GetUsersSubscriptions retrieves all subscriptions for multiple users in a single query
func (r *subscriptionRepository) GetUsersSubscriptions(userIDs []uint) ([]*models.Subscription, error) {
	if len(userIDs) == 0 {
		return []*models.Subscription{}, nil
	}
	var subscriptions []*models.Subscription
	err := r.db.Where("user_id IN ?", userIDs).Order("created_at DESC").Find(&subscriptions).Error
	return subscriptions, err
}

// GetActiveUserSubscriptions retrieves active subscriptions for a user
func (r *subscriptionRepository) GetActiveUserSubscriptions(userID uint) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.Where("user_id = ? AND status = ?", userID, models.SubscriptionStatusActive).
		Order("created_at DESC").Find(&subscriptions).Error
	return subscriptions, err
}

// GetUserTierSubscription retrieves the user's tier subscription
func (r *subscriptionRepository) GetUserTierSubscription(userID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Where("user_id = ? AND subscription_type = ? AND status = ?",
		userID, models.SubscriptionTypeTier, models.SubscriptionStatusActive).
		First(&subscription).Error
	return &subscription, err
}

// GetUserAddonSubscriptions retrieves all active addon subscriptions for a user
func (r *subscriptionRepository) GetUserAddonSubscriptions(userID uint) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.Where("user_id = ? AND subscription_type = ? AND status = ?",
		userID, models.SubscriptionTypeAddon, models.SubscriptionStatusActive).
		Order("created_at DESC").Find(&subscriptions).Error
	return subscriptions, err
}

// Update updates a subscription record
func (r *subscriptionRepository) Update(subscription *models.Subscription) error {
	return r.db.Save(subscription).Error
}

// Delete soft deletes a subscription
func (r *subscriptionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Subscription{}, id).Error
}

// GetAll retrieves all subscriptions with optional filters
func (r *subscriptionRepository) GetAll(status string, subscriptionType string, limit, offset int) ([]*models.Subscription, int64, error) {
	var subscriptions []*models.Subscription
	var total int64

	query := r.db.Model(&models.Subscription{}).Preload("User")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if subscriptionType != "" {
		query = query.Where("subscription_type = ?", subscriptionType)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error

	return subscriptions, total, err
}
