package stripe

import "github.com/stripe/stripe-go/v83"

// StripeClient is the unified interface for all Stripe operations
// All services should use this interface to interact with Stripe
type StripeClient interface {
	// Customer operations
	CreateCustomer(email, name string, metadata map[string]string) (*stripe.Customer, error)
	GetCustomer(customerID string) (*stripe.Customer, error)
	UpdateCustomer(customerID string, params *stripe.CustomerParams) (*stripe.Customer, error)

	// Checkout and billing portal
	CreateCheckoutSession(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error)
	GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error)
	CreateBillingPortalSession(customerID, returnURL string, configurationID ...string) (*stripe.BillingPortalSession, error)

	// Subscription management
	GetSubscription(subscriptionID string) (*stripe.Subscription, error)
	CancelSubscription(subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error)

	// Payment operations
	GetPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error)
	CreateRefund(paymentIntentID string, amount int64, reason string) (*stripe.Refund, error)

	// Coupon operations
	CreateCoupon(code string, percentageOff float64, maxRedemptions int) (*stripe.Coupon, error)
	GetCoupon(couponID string) (*stripe.Coupon, error)

	// Webhook verification
	ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error)
	VerifyWebhookSignature(payload []byte, signature string) error
}
