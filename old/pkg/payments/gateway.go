package payments

// PaymentService defines the interface for payment operations
// All payment operations are handled through Stripe, which supports both fiat (card) and crypto payments
type PaymentService interface {
	// CreateOrderPaymentLink creates a Stripe Checkout session for marketplace orders
	// Supports both card and cryptocurrency payments through Stripe
	CreateOrderPaymentLink(orderID uint, amount float64, description string) (*PaymentOrderLink, error)

	// CreateOrderCheckoutSession creates a Stripe Checkout session with full configuration
	CreateOrderCheckoutSession(params *OrderCheckoutParams) (*PaymentOrderLink, error)

	// GetPaymentStatus retrieves the current status of a payment
	GetPaymentStatus(paymentID string) (string, error)

	// CreateRefund issues a refund for a completed payment
	CreateRefund(paymentIntentID string, amount int64, reason string) error
}
