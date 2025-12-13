package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
)

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByID(id uint) (*models.Payment, error)
	GetByStripePaymentIntentID(paymentIntentID string) (*models.Payment, error)
	GetByStripeCheckoutSessionID(sessionID string) (*models.Payment, error)
	GetByStripeChargeID(chargeID string) (*models.Payment, error)
	GetByRelatedEntity(entityType string, entityID uint) (*models.Payment, error)
	GetUserPayments(userID uint, limit, offset int) ([]*models.Payment, int64, error)
	GetByProvider(provider models.PaymentProvider, limit, offset int) ([]*models.Payment, int64, error)
	GetByStatus(status models.PaymentStatus, limit, offset int) ([]*models.Payment, int64, error)
	Update(payment *models.Payment) error
	GetAll(filters PaymentFilters, limit, offset int) ([]*models.Payment, int64, error)
}

type PaymentFilters struct {
	Provider    models.PaymentProvider
	Status      models.PaymentStatus
	PaymentType models.PaymentType
	UserID      uint
}

type paymentRepository struct {
	*Repository
}

func NewPaymentRepository(r *Repository) PaymentRepository {
	return &paymentRepository{
		Repository: r,
	}
}

// Create creates a new payment record
func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// GetByID retrieves a payment by ID
func (r *paymentRepository) GetByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").First(&payment, id).Error
	return &payment, err
}

// GetByStripePaymentIntentID retrieves a payment by Stripe payment intent ID
func (r *paymentRepository) GetByStripePaymentIntentID(paymentIntentID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").
		Where("stripe_payment_intent_id = ?", paymentIntentID).
		First(&payment).Error
	return &payment, err
}

// GetByStripeCheckoutSessionID retrieves a payment by Stripe checkout session ID
func (r *paymentRepository) GetByStripeCheckoutSessionID(sessionID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").
		Where("stripe_checkout_session_id = ?", sessionID).
		First(&payment).Error
	return &payment, err
}

// GetByStripeChargeID retrieves a payment by Stripe charge ID
func (r *paymentRepository) GetByStripeChargeID(chargeID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").
		Where("stripe_charge_id = ?", chargeID).
		First(&payment).Error
	return &payment, err
}

// GetByRelatedEntity retrieves a payment for a specific entity
func (r *paymentRepository) GetByRelatedEntity(entityType string, entityID uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("related_entity_type = ? AND related_entity_id = ?", entityType, entityID).
		First(&payment).Error
	return &payment, err
}

// GetUserPayments retrieves all payments for a user
func (r *paymentRepository) GetUserPayments(userID uint, limit, offset int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Where("user_id = ?", userID)

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error

	return payments, total, err
}

// GetByProvider retrieves payments by provider
func (r *paymentRepository) GetByProvider(provider models.PaymentProvider, limit, offset int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Where("provider = ?", provider).Preload("User")

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error

	return payments, total, err
}

// GetByStatus retrieves payments by status
func (r *paymentRepository) GetByStatus(status models.PaymentStatus, limit, offset int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Where("status = ?", status).Preload("User")

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error

	return payments, total, err
}

// Update updates a payment record
func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

// GetAll retrieves all payments with optional filters
func (r *paymentRepository) GetAll(filters PaymentFilters, limit, offset int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Preload("User")

	if filters.Provider != "" {
		query = query.Where("provider = ?", filters.Provider)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.PaymentType != "" {
		query = query.Where("payment_type = ?", filters.PaymentType)
	}

	if filters.UserID > 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error

	return payments, total, err
}
