package payments

import (
	"fmt"

	"github.com/stripe/stripe-go/v83"
	checkoutsession "github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/refund"
)

// stripePaymentService implements PaymentService using Stripe
// Supports both fiat (card) and cryptocurrency payments through Stripe
type stripePaymentService struct {
	secretKey string
}

// NewStripePaymentGateway creates a new Stripe-based payment service
// All payments (including crypto) are processed through Stripe
func NewStripePaymentGateway(secretKey string) PaymentService {
	stripe.Key = secretKey
	return &stripePaymentService{
		secretKey: secretKey,
	}
}

// CreateOrderPaymentLink creates a Stripe Checkout session for marketplace orders
// This is a simplified version that uses default success/cancel URLs
func (s *stripePaymentService) CreateOrderPaymentLink(orderID uint, amount float64, description string) (*PaymentOrderLink, error) {
	params := &OrderCheckoutParams{
		OrderID:     orderID,
		Amount:      amount,
		Currency:    "usd",
		Description: description,
		SuccessURL:  fmt.Sprintf("https://app.hellobutter.io/orders/%d/success", orderID),
		CancelURL:   "https://app.hellobutter.io/marketplace",
	}

	return s.CreateOrderCheckoutSession(params)
}

// CreateOrderCheckoutSession creates a Stripe Checkout session with full configuration
// Supports both card and cryptocurrency payment methods
func (s *stripePaymentService) CreateOrderCheckoutSession(params *OrderCheckoutParams) (*PaymentOrderLink, error) {
	// Convert amount to cents (Stripe uses smallest currency unit)
	amountCents := int64(params.Amount * 100)

	// Create checkout session parameters
	checkoutParams := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(params.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(params.Description),
						Description: stripe.String(fmt.Sprintf("Order #%d", params.OrderID)),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),

		// Enable both card and crypto payment methods
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
			"crypto",
		}),

		// Add metadata for webhook processing
		Metadata: map[string]string{
			"order_id":     fmt.Sprintf("%d", params.OrderID),
			"user_id":      fmt.Sprintf("%d", params.UserID),
			"payment_type": "marketplace_order",
		},
	}

	// Set customer email if provided
	if params.UserEmail != "" {
		checkoutParams.CustomerEmail = stripe.String(params.UserEmail)
	}

	// Create the checkout session
	session, err := checkoutsession.New(checkoutParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe Checkout session: %w", err)
	}

	return &PaymentOrderLink{
		PaymentID:       session.ID,
		Url:             session.URL,
		Token:           session.ID, // For backwards compatibility
		CheckoutSession: session.ID,
	}, nil
}

// GetPaymentStatus retrieves the current status of a payment from Stripe
func (s *stripePaymentService) GetPaymentStatus(paymentID string) (string, error) {
	// Check if it's a checkout session ID
	session, err := checkoutsession.Get(paymentID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get checkout session: %w", err)
	}

	// Map Stripe status to our internal status
	switch session.PaymentStatus {
	case stripe.CheckoutSessionPaymentStatusPaid:
		return "completed", nil
	case stripe.CheckoutSessionPaymentStatusUnpaid:
		return "pending", nil
	case stripe.CheckoutSessionPaymentStatusNoPaymentRequired:
		return "completed", nil
	default:
		return "pending", nil
	}
}

// CreateRefund issues a refund for a completed payment
func (s *stripePaymentService) CreateRefund(paymentIntentID string, amount int64, reason string) error {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	// If amount is specified, do a partial refund
	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	// Add reason if provided
	if reason != "" {
		params.Reason = stripe.String(reason)
	}

	_, err := refund.New(params)
	if err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}

	return nil
}
