package stripe

import (
	"encoding/json"
	"time"

	"github.com/stripe/stripe-go/v83"
	billingportalsession "github.com/stripe/stripe-go/v83/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/coupon"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/paymentintent"
	"github.com/stripe/stripe-go/v83/refund"
	subscriptionClient "github.com/stripe/stripe-go/v83/subscription"
	"github.com/stripe/stripe-go/v83/webhook"
)

// Client wraps Stripe SDK functionality
type Client struct {
	secretKey     string
	webhookSecret string
}

// NewClient creates a new Stripe client
func NewClient(secretKey, webhookSecret string) *Client {
	stripe.Key = secretKey
	return &Client{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// CreateCustomer creates a new Stripe customer
func (c *Client) CreateCustomer(email, name string, metadata map[string]string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	if metadata != nil {
		for key, value := range metadata {
			params.AddMetadata(key, value)
		}
	}

	return customer.New(params)
}

// GetCustomer retrieves a customer by ID
func (c *Client) GetCustomer(customerID string) (*stripe.Customer, error) {
	return customer.Get(customerID, nil)
}

// UpdateCustomer updates a customer's information
func (c *Client) UpdateCustomer(customerID string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	return customer.Update(customerID, params)
}

// CreateCheckoutSession creates a new checkout session for subscriptions or one-time payments
func (c *Client) CreateCheckoutSession(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	return checkoutsession.New(params)
}

// GetCheckoutSession retrieves a checkout session by ID
func (c *Client) GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	return checkoutsession.Get(sessionID, nil)
}

// CreateBillingPortalSession creates a customer portal session
func (c *Client) CreateBillingPortalSession(customerID, returnURL string, configurationID ...string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	// Add configuration ID if provided
	if len(configurationID) > 0 && configurationID[0] != "" {
		params.Configuration = stripe.String(configurationID[0])
	}

	return billingportalsession.New(params)
}

// GetSubscription retrieves a subscription by ID
func (c *Client) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return subscriptionClient.Get(subscriptionID, nil)
}

// CancelSubscription cancels a subscription
func (c *Client) CancelSubscription(subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{}

	if cancelAtPeriodEnd {
		// Schedule cancellation at period end
		updateParams := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		return subscriptionClient.Update(subscriptionID, updateParams)
	}

	// Cancel immediately
	return subscriptionClient.Cancel(subscriptionID, params)
}

// GetPaymentIntent retrieves a payment intent by ID
func (c *Client) GetPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(paymentIntentID, nil)
}

// CreateRefund creates a refund for a payment
func (c *Client) CreateRefund(paymentIntentID string, amount int64, reason string) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	if reason != "" {
		params.Reason = stripe.String(reason)
	}

	return refund.New(params)
}

// ConstructWebhookEvent verifies and constructs a webhook event from the request
func (c *Client) ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, c.webhookSecret)
}

// VerifyWebhookSignature verifies the webhook signature
func (c *Client) VerifyWebhookSignature(payload []byte, signature string) error {
	_, err := webhook.ConstructEvent(payload, signature, c.webhookSecret)
	return err
}

// CreateCoupon creates a coupon with a percentage discount
// The code parameter is used as both the coupon ID and name for idempotency
// If a coupon with the same ID already exists, it retrieves and returns the existing coupon
func (c *Client) CreateCoupon(code string, percentageOff float64, maxRedemptions int) (*stripe.Coupon, error) {
	params := &stripe.CouponParams{
		ID:         stripe.String(code),
		Duration:   stripe.String(string(stripe.CouponDurationOnce)),
		PercentOff: stripe.Float64(percentageOff),
		Name:       stripe.String(code),
	}

	if maxRedemptions > 0 {
		params.MaxRedemptions = stripe.Int64(int64(maxRedemptions))
	}

	newCoupon, err := coupon.New(params)
	if err != nil {
		// Check if error is due to duplicate ID
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceAlreadyExists {
			// Coupon already exists, retrieve it
			return coupon.Get(code, nil)
		}
		return nil, err
	}

	return newCoupon, nil
}

// GetCoupon retrieves a coupon by ID
func (c *Client) GetCoupon(couponID string) (*stripe.Coupon, error) {
	return coupon.Get(couponID, nil)
}

// Helper functions for working with Stripe objects

// SafeGetSubscriptionID safely extracts subscription ID from CheckoutSession
// Handles both expanded and non-expanded Subscription objects
func SafeGetSubscriptionID(session *stripe.CheckoutSession) string {
	if session == nil || session.Subscription == nil {
		return ""
	}
	return session.Subscription.ID
}

// SafeGetPaymentIntentID safely extracts payment intent ID from CheckoutSession
// Handles both expanded and non-expanded PaymentIntent objects
func SafeGetPaymentIntentID(session *stripe.CheckoutSession) string {
	if session == nil || session.PaymentIntent == nil {
		return ""
	}
	return session.PaymentIntent.ID
}

// SafeGetCustomerID safely extracts customer ID from expandable Customer
// Handles both expanded and non-expanded Customer objects
func SafeGetCustomerID(customer *stripe.Customer) string {
	if customer == nil {
		return ""
	}
	return customer.ID
}

// GetPaymentMethodID extracts payment method ID from subscription
func GetPaymentMethodID(subscription *stripe.Subscription) string {
	if subscription == nil {
		return ""
	}
	if subscription.DefaultPaymentMethod != nil {
		return subscription.DefaultPaymentMethod.ID
	}
	return ""
}

// GetCurrentPeriodFromSubscription extracts current period dates from subscription
// In v83, CurrentPeriodStart/End were removed from Subscription
// We extract from LatestInvoice.PeriodStart/End if available, otherwise use defaults
func GetCurrentPeriodFromSubscription(sub *stripe.Subscription) (start time.Time, end time.Time) {
	// Default to epoch time if we can't determine
	start = time.Unix(0, 0)
	end = time.Unix(0, 0)

	if sub == nil {
		return
	}

	// Try to get period from latest invoice
	if sub.LatestInvoice != nil {
		if sub.LatestInvoice.PeriodStart > 0 {
			start = time.Unix(sub.LatestInvoice.PeriodStart, 0)
		}
		if sub.LatestInvoice.PeriodEnd > 0 {
			end = time.Unix(sub.LatestInvoice.PeriodEnd, 0)
		}
	}

	return
}

// ExtractSubscriptionIDFromInvoice extracts subscription ID from invoice
// In stripe-go v83, Subscription field is not directly accessible on Invoice struct
// We must parse from raw JSON response
func ExtractSubscriptionIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice == nil {
		return ""
	}

	// Try to get from LastResponse RawJSON if available
	if invoice.LastResponse != nil && len(invoice.LastResponse.RawJSON) > 0 {
		var rawData map[string]interface{}
		if err := json.Unmarshal(invoice.LastResponse.RawJSON, &rawData); err == nil {
			if subID, ok := rawData["subscription"].(string); ok && subID != "" {
				return subID
			}
		}
	}

	return ""
}

// ExtractPaymentIntentIDFromInvoice extracts payment intent ID from invoice
// In stripe-go v83, PaymentIntent field is not directly accessible on Invoice struct
// We must parse from raw JSON response
func ExtractPaymentIntentIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice == nil {
		return ""
	}

	// Try to get from LastResponse RawJSON if available
	if invoice.LastResponse != nil && len(invoice.LastResponse.RawJSON) > 0 {
		var rawData map[string]interface{}
		if err := json.Unmarshal(invoice.LastResponse.RawJSON, &rawData); err == nil {
			if piID, ok := rawData["payment_intent"].(string); ok && piID != "" {
				return piID
			}
		}
	}

	return ""
}

// IsCryptoPaymentMethod checks if a payment method type is crypto-related
func IsCryptoPaymentMethod(paymentMethodType string) bool {
	cryptoTypes := []string{"bitcoin", "ethereum", "usdc", "usdt", "dai"}
	for _, cryptoType := range cryptoTypes {
		if paymentMethodType == cryptoType {
			return true
		}
	}
	return false
}
