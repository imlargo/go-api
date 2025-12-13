package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
	stripeClient "github.com/nicolailuther/butter/pkg/stripe"
	"github.com/stripe/stripe-go/v83"
)

type WebhookHandler struct {
	*Handler
	subscriptionService services.SubscriptionService
	marketplaceService  services.MarketplaceService
	paymentService      services.PaymentService
	stripeClient        stripeClient.StripeClient
}

func NewWebhookHandler(
	h *Handler,
	subscriptionService services.SubscriptionService,
	marketplaceService services.MarketplaceService,
	paymentService services.PaymentService,
	stripeClient stripeClient.StripeClient,
) *WebhookHandler {
	return &WebhookHandler{
		Handler:             h,
		subscriptionService: subscriptionService,
		marketplaceService:  marketplaceService,
		paymentService:      paymentService,
		stripeClient:        stripeClient,
	}
}

// HandleStripeWebhook processes Stripe webhook events
// @Summary Handle Stripe webhooks
// @Description Process incoming Stripe webhook events for payments and subscriptions
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/webhooks/stripe [post]
func (h *WebhookHandler) HandleStripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Errorf("Error reading request body: %v", err)
		responses.ErrorBadRequest(c, "Error reading request body")
		return
	}

	// Verify webhook signature
	signature := c.GetHeader("Stripe-Signature")
	event, err := h.stripeClient.ConstructWebhookEvent(payload, signature)
	if err != nil {
		h.logger.Errorf("Error verifying webhook signature: %v", err)
		responses.ErrorBadRequest(c, "Invalid signature")
		return
	}

	h.logger.Infof("Received Stripe webhook event: %s (ID: %s)", event.Type, event.ID)

	// Handle the event based on type
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			h.logger.Errorf("Error unmarshaling checkout session for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}

		if session.Mode == stripe.CheckoutSessionModeSubscription {
			if err := h.subscriptionService.HandleCheckoutSessionCompleted(&session); err != nil {
				h.logger.Errorf("Error handling checkout session completed (session: %s, event: %s): %v", session.ID, event.ID, err)
				responses.ErrorInternalServerWithMessage(c, "Error processing event")
				return
			}
			h.logger.Infof("Successfully processed checkout.session.completed for subscription (session: %s)", session.ID)
		} else {
			// Marketplace payment
			if err := h.handleMarketplacePaymentCompleted(&session); err != nil {
				h.logger.Errorf("Error handling marketplace payment completed (session: %s, event: %s): %v", session.ID, event.ID, err)
				responses.ErrorInternalServerWithMessage(c, "Error processing event")
				return
			}
			h.logger.Infof("Successfully processed checkout.session.completed for marketplace (session: %s)", session.ID)
		}

	case "customer.subscription.created":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			h.logger.Errorf("Error unmarshaling subscription for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.subscriptionService.HandleSubscriptionCreated(&subscription); err != nil {
			h.logger.Errorf("Error handling subscription created (subscription: %s, event: %s): %v", subscription.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed customer.subscription.created (subscription: %s)", subscription.ID)

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			h.logger.Errorf("Error unmarshaling subscription for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.subscriptionService.HandleSubscriptionUpdated(&subscription); err != nil {
			h.logger.Errorf("Error handling subscription updated (subscription: %s, event: %s): %v", subscription.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed customer.subscription.updated (subscription: %s)", subscription.ID)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			h.logger.Errorf("Error unmarshaling subscription for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.subscriptionService.HandleSubscriptionDeleted(&subscription); err != nil {
			h.logger.Errorf("Error handling subscription deleted (subscription: %s, event: %s): %v", subscription.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed customer.subscription.deleted (subscription: %s)", subscription.ID)

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			h.logger.Errorf("Error unmarshaling invoice for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.subscriptionService.HandleInvoicePaymentSucceeded(&invoice); err != nil {
			h.logger.Errorf("Error handling invoice payment succeeded (invoice: %s, event: %s): %v", invoice.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed invoice.payment_succeeded (invoice: %s)", invoice.ID)

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			h.logger.Errorf("Error unmarshaling invoice for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.subscriptionService.HandleInvoicePaymentFailed(&invoice); err != nil {
			h.logger.Errorf("Error handling invoice payment failed (invoice: %s, event: %s): %v", invoice.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed invoice.payment_failed (invoice: %s)", invoice.ID)

	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			h.logger.Errorf("Error unmarshaling payment intent for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.handlePaymentIntentSucceeded(&paymentIntent); err != nil {
			h.logger.Errorf("Error handling payment intent succeeded (payment_intent: %s, event: %s): %v", paymentIntent.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed payment_intent.succeeded (payment_intent: %s)", paymentIntent.ID)

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			h.logger.Errorf("Error unmarshaling payment intent for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.handlePaymentIntentFailed(&paymentIntent); err != nil {
			h.logger.Errorf("Error handling payment intent failed (payment_intent: %s, event: %s): %v", paymentIntent.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed payment_intent.payment_failed (payment_intent: %s)", paymentIntent.ID)

	case "payment_intent.canceled":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			h.logger.Errorf("Error unmarshaling payment intent for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.handlePaymentIntentCanceled(&paymentIntent); err != nil {
			h.logger.Errorf("Error handling payment intent canceled (payment_intent: %s, event: %s): %v", paymentIntent.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed payment_intent.canceled (payment_intent: %s)", paymentIntent.ID)

	case "charge.succeeded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			h.logger.Errorf("Error unmarshaling charge for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.handleChargeSucceeded(&charge); err != nil {
			h.logger.Errorf("Error handling charge succeeded (charge: %s, event: %s): %v", charge.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed charge.succeeded (charge: %s)", charge.ID)

	case "charge.failed":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			h.logger.Errorf("Error unmarshaling charge for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.handleChargeFailed(&charge); err != nil {
			h.logger.Errorf("Error handling charge failed (charge: %s, event: %s): %v", charge.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed charge.failed (charge: %s)", charge.ID)

	case "charge.refunded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			h.logger.Errorf("Error unmarshaling charge for event %s: %v", event.ID, err)
			responses.ErrorBadRequest(c, "Error parsing event data")
			return
		}
		if err := h.handleChargeRefunded(&charge); err != nil {
			h.logger.Errorf("Error handling charge refunded (charge: %s, event: %s): %v", charge.ID, event.ID, err)
			responses.ErrorInternalServerWithMessage(c, "Error processing event")
			return
		}
		h.logger.Infof("Successfully processed charge.refunded (charge: %s)", charge.ID)

	default:
		h.logger.Infof("Unhandled event type: %s (event: %s)", event.Type, event.ID)
	}

	responses.Ok(c, gin.H{"received": true})
}

// handleMarketplacePaymentCompleted processes marketplace payment completion
func (h *WebhookHandler) handleMarketplacePaymentCompleted(session *stripe.CheckoutSession) error {
	// Verify this is a marketplace payment by checking metadata
	if session.Metadata["payment_type"] != "marketplace_order" {
		h.logger.Warnf("Checkout session %s is not a marketplace payment", session.ID)
		return nil
	}

	// Call marketplace service to handle the payment
	if err := h.marketplaceService.HandleOrderPaymentCompleted(session.ID); err != nil {
		return fmt.Errorf("failed to handle marketplace payment: %w", err)
	}

	h.logger.Infof("Marketplace payment completed successfully: session %s", session.ID)
	return nil
}

// handlePaymentIntentSucceeded updates payment records when a payment intent succeeds
func (h *WebhookHandler) handlePaymentIntentSucceeded(paymentIntent *stripe.PaymentIntent) error {
	return h.paymentService.HandlePaymentIntentSucceeded(paymentIntent)
}

// handlePaymentIntentFailed updates payment records when a payment intent fails
func (h *WebhookHandler) handlePaymentIntentFailed(paymentIntent *stripe.PaymentIntent) error {
	return h.paymentService.HandlePaymentIntentFailed(paymentIntent)
}

// handlePaymentIntentCanceled updates payment records when a payment intent is canceled
func (h *WebhookHandler) handlePaymentIntentCanceled(paymentIntent *stripe.PaymentIntent) error {
	return h.paymentService.HandlePaymentIntentCanceled(paymentIntent)
}

// handleChargeSucceeded updates payment records when a charge succeeds
func (h *WebhookHandler) handleChargeSucceeded(charge *stripe.Charge) error {
	return h.paymentService.HandleChargeSucceeded(charge)
}

// handleChargeFailed updates payment records when a charge fails
func (h *WebhookHandler) handleChargeFailed(charge *stripe.Charge) error {
	return h.paymentService.HandleChargeFailed(charge)
}

// handleChargeRefunded updates payment records when a charge is refunded
func (h *WebhookHandler) handleChargeRefunded(charge *stripe.Charge) error {
	return h.paymentService.HandleChargeRefunded(charge)
}
