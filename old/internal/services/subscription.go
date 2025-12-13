package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	stripeClient "github.com/nicolailuther/butter/pkg/stripe"
	"github.com/stripe/stripe-go/v83"
	"gorm.io/gorm"
)

type SubscriptionService interface {
	// Subscription management
	CreateCheckoutSession(userID uint, priceID string, successURL, cancelURL string) (string, error)
	CreateCustomerPortalSession(userID uint, returnURL string) (string, error)
	GetUserSubscriptions(userID uint) ([]*models.Subscription, error)
	GetUserTierSubscription(userID uint) (*models.Subscription, error)
	GetAvailablePlans() ([]SubscriptionPlan, error)

	// Webhook handlers
	HandleCheckoutSessionCompleted(session *stripe.CheckoutSession) error
	HandleSubscriptionCreated(subscription *stripe.Subscription) error
	HandleSubscriptionUpdated(subscription *stripe.Subscription) error
	HandleSubscriptionDeleted(subscription *stripe.Subscription) error
	HandleInvoicePaymentSucceeded(invoice *stripe.Invoice) error
	HandleInvoicePaymentFailed(invoice *stripe.Invoice) error

	// Admin functions
	GetAllSubscriptions(status string, subscriptionType string, page, pageSize int) ([]*models.Subscription, int64, error)
}

type SubscriptionPlan struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceID     string   `json:"price_id"`
	Price       float64  `json:"price"` // in cents
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"`
	Features    []string `json:"features"`
}

type subscriptionService struct {
	*Service
	stripeClient stripeClient.StripeClient
}

func NewSubscriptionService(s *Service, stripeClient stripeClient.StripeClient) SubscriptionService {
	return &subscriptionService{
		Service:      s,
		stripeClient: stripeClient,
	}
}

// CreateCheckoutSession creates a Stripe Checkout session for a subscription
func (s *subscriptionService) CreateCheckoutSession(userID uint, priceID string, successURL, cancelURL string) (string, error) {
	// Get user
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Get or create Stripe customer
	var customerID string
	if user.StripeCustomerID != "" {
		customerID = user.StripeCustomerID
	} else {
		customer, err := s.stripeClient.CreateCustomer(user.Email, user.Name, map[string]string{
			"user_id": fmt.Sprintf("%d", user.ID),
		})
		if err != nil {
			return "", fmt.Errorf("failed to create Stripe customer: %w", err)
		}
		customerID = customer.ID

		// Update user with Stripe customer ID
		user.StripeCustomerID = customerID
		if err := s.store.Users.Update(user); err != nil {
			s.logger.Warnf("Failed to save Stripe customer ID for user %d: %v", userID, err)
		}
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}
	params.AddMetadata("user_id", fmt.Sprintf("%d", userID))

	// Check if user has a referral code with discount and hasn't used it yet
	if user.ReferralCodeID > 0 {
		hasUsedDiscount, err := s.store.ReferralDiscounts.HasUserUsedDiscount(userID)
		if err != nil {
			s.logger.Warnf("Failed to check if user %d has used discount: %v", userID, err)
		} else if !hasUsedDiscount {
			referralCode, err := s.store.ReferralCode.GetByID(user.ReferralCodeID)
			if err != nil {
				s.logger.Warnf("Failed to get referral code %d: %v", user.ReferralCodeID, err)
			} else if referralCode.DiscountPercentage != nil {
				// Create or get Stripe coupon
				couponID := referralCode.StripeCouponID
				if couponID == "" {
					// Create new coupon in Stripe with prefixed ID
					coupon, err := s.stripeClient.CreateCoupon(
						"ref_"+referralCode.Code,
						*referralCode.DiscountPercentage,
						0, // No max redemptions for the coupon itself
					)
					if err == nil {
						couponID = coupon.ID
						// Update referral code with Stripe coupon ID
						referralCode.StripeCouponID = couponID
						if updateErr := s.store.ReferralCode.Update(referralCode); updateErr != nil {
							s.logger.Errorf("Failed to update referral code %d with Stripe coupon ID: %v", referralCode.ID, updateErr)
						}
					} else {
						s.logger.Errorf("Failed to create Stripe coupon for referral code %s: %v - checkout will proceed without discount", referralCode.Code, err)
					}
				}

				// Apply discount to checkout session
				if couponID != "" {
					params.Discounts = []*stripe.CheckoutSessionDiscountParams{
						{
							Coupon: stripe.String(couponID),
						},
					}
					params.AddMetadata("referral_code_id", fmt.Sprintf("%d", referralCode.ID))
					params.AddMetadata("discount_percentage", fmt.Sprintf("%.2f", *referralCode.DiscountPercentage))
					s.logger.Infof("Applied %.2f%% discount to checkout session for user %d using referral code %s", *referralCode.DiscountPercentage, userID, referralCode.Code)
				}
			}
		}
	}

	session, err := s.stripeClient.CreateCheckoutSession(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return session.URL, nil
}

// CreateCustomerPortalSession creates a Stripe Customer Portal session
func (s *subscriptionService) CreateCustomerPortalSession(userID uint, returnURL string) (string, error) {
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user.StripeCustomerID == "" {
		return "", errors.New("user does not have a Stripe customer ID")
	}

	// Pass configuration ID if available
	var session *stripe.BillingPortalSession
	if s.config.Stripe.PortalConfigurationID != "" {
		session, err = s.stripeClient.CreateBillingPortalSession(user.StripeCustomerID, returnURL, s.config.Stripe.PortalConfigurationID)
	} else {
		session, err = s.stripeClient.CreateBillingPortalSession(user.StripeCustomerID, returnURL)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create billing portal session: %w", err)
	}

	return session.URL, nil
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (s *subscriptionService) GetUserSubscriptions(userID uint) ([]*models.Subscription, error) {
	return s.store.Subscriptions.GetUserSubscriptions(userID)
}

// GetUserTierSubscription retrieves the user's tier subscription
func (s *subscriptionService) GetUserTierSubscription(userID uint) (*models.Subscription, error) {
	return s.store.Subscriptions.GetUserTierSubscription(userID)
}

// GetAvailablePlans returns the available subscription plans
func (s *subscriptionService) GetAvailablePlans() ([]SubscriptionPlan, error) {
	plans := []SubscriptionPlan{
		{
			Name:        "Free",
			Description: "Basic platform access",
			PriceID:     "",
			Price:       0,
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"Basic access",
				"Marketplace done for you services",
			},
		},
		{
			Name:        "Starter",
			Description: "Enhanced features for growing creators",
			PriceID:     s.config.Stripe.PriceStarter,
			Price:       199.99, // $29/month
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"Everything in Free",
				"5 Social Accounts",
				"Basic Analytics",
				"Included Posting CRM",
			},
		},
		{
			Name:        "Growth",
			Description: "For growing teams and creators",
			PriceID:     s.config.Stripe.PriceGrowth,
			Price:       299.99, // $79/month
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"Everything in Starter",
				"Up to 10 Social Accounts",
				"Complete Analytics",
				"Team members",
			},
		},
		{
			Name:        "Scale",
			Description: "For agencies and power users",
			PriceID:     s.config.Stripe.PriceScale,
			Price:       399.99, // $199/month
			Currency:    "usd",
			Interval:    "month",
			Features: []string{
				"Everything in Growth",
				"20 Social Accounts",
				"Included Posting CRM",
				"Unlimited Team members",
			},
		},
	}

	return plans, nil
}

// HandleCheckoutSessionCompleted processes checkout.session.completed webhook
func (s *subscriptionService) HandleCheckoutSessionCompleted(session *stripe.CheckoutSession) error {
	if session.Mode != stripe.CheckoutSessionModeSubscription {
		// This is a marketplace payment, not a subscription
		return nil
	}

	userIDStr, ok := session.Metadata["user_id"]
	if !ok {
		return errors.New("user_id not found in session metadata")
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	// Safely extract subscription ID from expandable object
	subscriptionID := stripeClient.SafeGetSubscriptionID(session)
	if subscriptionID == "" {
		s.logger.Warnf("Processing checkout session completed for user %d, but subscription ID not available", userID)
	} else {
		s.logger.Infof("Processing checkout session completed for user %d, subscription %s", userID, subscriptionID)
	}

	// The subscription will be created via the subscription.created webhook
	return nil
}

// HandleSubscriptionCreated processes customer.subscription.created webhook
func (s *subscriptionService) HandleSubscriptionCreated(stripeSubscription *stripe.Subscription) error {
	// Extract customer ID safely from expandable Customer object
	customerID := stripeClient.SafeGetCustomerID(stripeSubscription.Customer)
	if customerID == "" {
		return errors.New("customer ID not found in subscription")
	}

	// Get customer to extract user ID from metadata
	customer, err := s.stripeClient.GetCustomer(customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer %s: %w", customerID, err)
	}

	userIDStr, ok := customer.Metadata["user_id"]
	if !ok {
		return errors.New("user_id not found in customer metadata")
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	// Validate subscription has items
	if len(stripeSubscription.Items.Data) == 0 {
		return errors.New("subscription has no items")
	}

	// Determine subscription type (tier or addon)
	subscriptionType := models.SubscriptionTypeTier
	priceID := stripeSubscription.Items.Data[0].Price.ID

	// If user already has an active tier subscription, this is an addon
	existingTier, err := s.store.Subscriptions.GetUserTierSubscription(userID)
	if err == nil && existingTier != nil && existingTier.StripeSubscriptionID != stripeSubscription.ID {
		subscriptionType = models.SubscriptionTypeAddon
	}

	// Extract current period from subscription (via LatestInvoice in v83)
	start, end := stripeClient.GetCurrentPeriodFromSubscription(stripeSubscription)

	// Create subscription record
	subscription := &models.Subscription{
		UserID:                 userID,
		StripeSubscriptionID:   stripeSubscription.ID,
		StripePriceID:          priceID,
		StripeProductID:        stripeSubscription.Items.Data[0].Price.Product.ID,
		SubscriptionType:       subscriptionType,
		TierLevel:              s.getTierLevelFromPriceID(priceID),
		Status:                 models.SubscriptionStatus(stripeSubscription.Status),
		CurrentPeriodStart:     start,
		CurrentPeriodEnd:       end,
		CancelAtPeriodEnd:      stripeSubscription.CancelAtPeriodEnd,
		DefaultPaymentMethodID: stripeClient.GetPaymentMethodID(stripeSubscription),
	}

	if stripeSubscription.CanceledAt > 0 {
		canceledAt := time.Unix(stripeSubscription.CanceledAt, 0)
		subscription.CanceledAt = &canceledAt
	}

	if err := s.store.Subscriptions.Create(subscription); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Check if a referral discount was applied and record it
	if len(stripeSubscription.Discounts) > 0 && stripeSubscription.Discounts[0].Source != nil && stripeSubscription.Discounts[0].Source.Coupon != nil {
		// Get user to check for referral code
		user, err := s.store.Users.GetByID(userID)
		if err == nil && user.ReferralCodeID > 0 {
			// Verify user hasn't used a discount before
			hasUsedDiscount, _ := s.store.ReferralDiscounts.HasUserUsedDiscount(userID)
			if !hasUsedDiscount {
				referralCode, err := s.store.ReferralCode.GetByID(user.ReferralCodeID)
				if err == nil && referralCode.DiscountPercentage != nil {
					appliedCouponID := stripeSubscription.Discounts[0].Source.Coupon.ID
					// Verify this discount matches the user's referral code
					if referralCode.StripeCouponID != "" && appliedCouponID == referralCode.StripeCouponID {
						// Record the discount usage
						discount := &models.ReferralDiscount{
							UserID:             userID,
							ReferralCodeID:     referralCode.ID,
							DiscountPercentage: *referralCode.DiscountPercentage,
							StripeCouponID:     appliedCouponID,
							SubscriptionID:     subscription.ID,
							AppliedAt:          time.Now(),
						}
						if err := s.store.ReferralDiscounts.Create(discount); err != nil {
							s.logger.Errorf("Failed to record referral discount usage: %v", err)
						} else {
							s.logger.Infof("Recorded referral discount usage for user %d: %.2f%%", userID, *referralCode.DiscountPercentage)
						}
					}
				}
			}
		}
	}

	// Update user tier if this is a tier subscription
	if subscriptionType == models.SubscriptionTypeTier {
		if err := s.updateUserTier(userID, priceID, string(stripeSubscription.Status)); err != nil {
			s.logger.Errorf("Failed to update user tier: %v", err)
		}
	}

	// Create payment record for the initial payment
	if stripeSubscription.LatestInvoice != nil {
		metadataJSON, _ := json.Marshal(map[string]interface{}{
			"subscription_id": stripeSubscription.ID,
			"price_id":        priceID,
		})

		now := time.Now()
		payment := &models.Payment{
			UserID:            userID,
			Provider:          models.PaymentProviderStripe,
			Status:            models.PaymentStatusCompleted,
			PaymentType:       models.PaymentTypeSubscription,
			Amount:            stripeSubscription.Items.Data[0].Price.UnitAmount,
			Currency:          string(stripeSubscription.Items.Data[0].Price.Currency),
			StripeInvoiceID:   stripeSubscription.LatestInvoice.ID,
			RelatedEntityType: "subscription",
			RelatedEntityID:   subscription.ID,
			ProcessedAt:       &now,
			CompletedAt:       &now,
			Metadata:          string(metadataJSON),
		}

		if err := s.store.Payments.Create(payment); err != nil {
			s.logger.Errorf("Failed to create payment record: %v", err)
		}
	}

	s.logger.Infof("Subscription created for user %d: %s", userID, stripeSubscription.ID)
	return nil
}

// HandleSubscriptionUpdated processes customer.subscription.updated webhook
func (s *subscriptionService) HandleSubscriptionUpdated(stripeSubscription *stripe.Subscription) error {
	subscription, err := s.store.Subscriptions.GetByStripeID(stripeSubscription.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Subscription doesn't exist yet, try to create it
			// If creation fails due to missing user_id in customer metadata, log warning and skip
			if createErr := s.HandleSubscriptionCreated(stripeSubscription); createErr != nil {
				s.logger.Warnf("Cannot create subscription %s during update: %v. Skipping update for untracked subscription.", stripeSubscription.ID, createErr)
				return nil // Don't fail the webhook - subscription might not be ours
			}
			return nil
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Extract current period from subscription (via LatestInvoice in v83)
	start, end := stripeClient.GetCurrentPeriodFromSubscription(stripeSubscription)

	// Get the price ID from the subscription items
	if len(stripeSubscription.Items.Data) == 0 {
		return errors.New("subscription has no items")
	}
	priceID := stripeSubscription.Items.Data[0].Price.ID

	// Update subscription fields (only if price ID changed to avoid unnecessary updates)
	if subscription.StripePriceID != priceID {
		subscription.StripePriceID = priceID
		subscription.TierLevel = s.getTierLevelFromPriceID(priceID)
	}
	subscription.Status = models.SubscriptionStatus(stripeSubscription.Status)
	subscription.CurrentPeriodStart = start
	subscription.CurrentPeriodEnd = end
	subscription.CancelAtPeriodEnd = stripeSubscription.CancelAtPeriodEnd
	subscription.DefaultPaymentMethodID = stripeClient.GetPaymentMethodID(stripeSubscription)

	if stripeSubscription.CanceledAt > 0 {
		canceledAt := time.Unix(stripeSubscription.CanceledAt, 0)
		subscription.CanceledAt = &canceledAt
	}

	if err := s.store.Subscriptions.Update(subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user tier if this is a tier subscription
	// IMPORTANT: Only update tier if subscription is in an appropriate state
	// Do NOT downgrade if cancel_at_period_end is true but subscription is still active
	if subscription.SubscriptionType == models.SubscriptionTypeTier {
		shouldUpdateTier := true

		// If subscription is canceled but still in the active period (cancel_at_period_end=true)
		// we should NOT downgrade the user until it actually ends
		if stripeSubscription.CancelAtPeriodEnd &&
			(stripeSubscription.Status == stripe.SubscriptionStatusActive ||
				stripeSubscription.Status == stripe.SubscriptionStatusTrialing) {
			s.logger.Infof("Subscription %s is set to cancel at period end but still active - keeping user tier", stripeSubscription.ID)
			shouldUpdateTier = false
		}

		// Only update tier for active, trialing, or past_due statuses
		// Don't downgrade for incomplete, incomplete_expired, unpaid unless they're actually canceled
		if shouldUpdateTier {
			if stripeSubscription.Status == stripe.SubscriptionStatusActive ||
				stripeSubscription.Status == stripe.SubscriptionStatusTrialing ||
				stripeSubscription.Status == stripe.SubscriptionStatusPastDue {
				if err := s.updateUserTier(subscription.UserID, subscription.StripePriceID, string(stripeSubscription.Status)); err != nil {
					s.logger.Errorf("Failed to update user tier: %v", err)
				}
			} else {
				s.logger.Infof("Subscription %s status is %s - not updating user tier", stripeSubscription.ID, stripeSubscription.Status)
			}
		}
	}

	s.logger.Infof("Subscription updated: %s (status: %s, cancel_at_period_end: %v)",
		stripeSubscription.ID, stripeSubscription.Status, stripeSubscription.CancelAtPeriodEnd)
	return nil
}

// HandleSubscriptionDeleted processes customer.subscription.deleted webhook
func (s *subscriptionService) HandleSubscriptionDeleted(stripeSubscription *stripe.Subscription) error {
	subscription, err := s.store.Subscriptions.GetByStripeID(stripeSubscription.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Subscription %s not found in database during deletion event", stripeSubscription.ID)
			return nil
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	subscription.Status = models.SubscriptionStatusCanceled
	if err := s.store.Subscriptions.Update(subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// If this is a tier subscription, downgrade user to free tier
	// This event is only fired when the subscription actually ends (not when cancel_at_period_end is set)
	if subscription.SubscriptionType == models.SubscriptionTypeTier {
		user, err := s.store.Users.GetByID(subscription.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		user.TierLevel = enums.TierFree
		user.SubscriptionStatus = "canceled"
		if err := s.store.Users.Update(user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		s.logger.Infof("User %d downgraded to free tier - subscription %s ended", subscription.UserID, stripeSubscription.ID)
	}

	s.logger.Infof("Subscription deleted: %s", stripeSubscription.ID)
	return nil
}

// HandleInvoicePaymentSucceeded processes invoice.payment_succeeded webhook
func (s *subscriptionService) HandleInvoicePaymentSucceeded(invoice *stripe.Invoice) error {
	// Extract subscription ID safely (expandable object in v83)
	subscriptionID := stripeClient.ExtractSubscriptionIDFromInvoice(invoice)
	if subscriptionID == "" {
		return nil // Not a subscription invoice
	}

	subscription, err := s.store.Subscriptions.GetByStripeID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Create payment record
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"invoice_id":      invoice.ID,
		"subscription_id": subscriptionID,
	})

	// Extract payment intent ID safely (expandable object in v83)
	paymentIntentID := stripeClient.ExtractPaymentIntentIDFromInvoice(invoice)

	now := time.Now()
	payment := &models.Payment{
		UserID:                subscription.UserID,
		Provider:              models.PaymentProviderStripe,
		Status:                models.PaymentStatusCompleted,
		PaymentType:           models.PaymentTypeSubscription,
		Amount:                invoice.AmountPaid,
		Currency:              string(invoice.Currency),
		StripePaymentIntentID: paymentIntentID,
		StripeInvoiceID:       invoice.ID,
		RelatedEntityType:     "subscription",
		RelatedEntityID:       subscription.ID,
		ProcessedAt:           &now,
		CompletedAt:           &now,
		Metadata:              string(metadataJSON),
	}

	if err := s.store.Payments.Create(payment); err != nil {
		s.logger.Errorf("Failed to create payment record: %v", err)
	}

	s.logger.Infof("Invoice payment succeeded for subscription %s", subscriptionID)
	return nil
}

// HandleInvoicePaymentFailed processes invoice.payment_failed webhook
func (s *subscriptionService) HandleInvoicePaymentFailed(invoice *stripe.Invoice) error {
	// Extract subscription ID safely (expandable object in v83)
	subscriptionID := stripeClient.ExtractSubscriptionIDFromInvoice(invoice)
	if subscriptionID == "" {
		return nil // Not a subscription invoice
	}

	subscription, err := s.store.Subscriptions.GetByStripeID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Create failed payment record
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"invoice_id":      invoice.ID,
		"subscription_id": subscriptionID,
	})

	// Extract payment intent ID and failure reason safely (expandable object in v83)
	paymentIntentID := stripeClient.ExtractPaymentIntentIDFromInvoice(invoice)
	failureReason := ""
	// Note: In v83, we can't directly access PaymentIntent.LastPaymentError from invoice
	// We would need to fetch the PaymentIntent separately if detailed error info is needed

	now := time.Now()
	payment := &models.Payment{
		UserID:                subscription.UserID,
		Provider:              models.PaymentProviderStripe,
		Status:                models.PaymentStatusFailed,
		PaymentType:           models.PaymentTypeSubscription,
		Amount:                invoice.AmountDue,
		Currency:              string(invoice.Currency),
		StripePaymentIntentID: paymentIntentID,
		StripeInvoiceID:       invoice.ID,
		RelatedEntityType:     "subscription",
		RelatedEntityID:       subscription.ID,
		FailureReason:         failureReason,
		ProcessedAt:           &now,
		Metadata:              string(metadataJSON),
	}

	if err := s.store.Payments.Create(payment); err != nil {
		s.logger.Errorf("Failed to create payment record: %v", err)
	}

	s.logger.Warnf("Invoice payment failed for subscription %s: %s", subscriptionID, failureReason)
	return nil
}

// GetAllSubscriptions retrieves all subscriptions with pagination
func (s *subscriptionService) GetAllSubscriptions(status string, subscriptionType string, page, pageSize int) ([]*models.Subscription, int64, error) {
	offset := (page - 1) * pageSize
	return s.store.Subscriptions.GetAll(status, subscriptionType, pageSize, offset)
}

// Helper functions

func (s *subscriptionService) getTierLevelFromPriceID(priceID string) enums.TierLevel {
	// Map price ID to tier level
	switch priceID {
	case s.config.Stripe.PriceStarter:
		return enums.TierStarter
	case s.config.Stripe.PriceGrowth:
		return enums.TierGrowth
	case s.config.Stripe.PriceScale:
		return enums.TierScale
	default:
		return enums.TierFree // Free tier
	}
}

func (s *subscriptionService) updateUserTier(userID uint, priceID string, status string) error {
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return err
	}

	user.TierLevel = s.getTierLevelFromPriceID(priceID)
	user.SubscriptionStatus = status

	return s.store.Users.Update(user)
}
