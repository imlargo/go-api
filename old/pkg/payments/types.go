package payments

// PaymentOrderLink contains the checkout URL and payment references for an order
type PaymentOrderLink struct {
	PaymentID       string `json:"payment_id"`       // Stripe Checkout Session ID
	Url             string `json:"url"`              // Stripe Checkout URL
	Token           string `json:"token"`            // Deprecated: kept for backwards compatibility
	CheckoutSession string `json:"checkout_session"` // Stripe Checkout Session ID
}

// OrderCheckoutParams contains parameters for creating an order checkout session
type OrderCheckoutParams struct {
	OrderID     uint    `json:"order_id"`
	Amount      float64 `json:"amount"`      // Amount in USD (or configured currency)
	Currency    string  `json:"currency"`    // Currency code (e.g., "usd")
	Description string  `json:"description"` // Order description
	UserID      uint    `json:"user_id"`     // Buyer user ID
	UserEmail   string  `json:"user_email"`  // Buyer email
	SuccessURL  string  `json:"success_url"` // URL to redirect after successful payment
	CancelURL   string  `json:"cancel_url"`  // URL to redirect if payment is cancelled
}
