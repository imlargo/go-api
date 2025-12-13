package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	stripeClient "github.com/nicolailuther/butter/pkg/stripe"
	"github.com/stripe/stripe-go/v83"
	"gorm.io/gorm"
)

type PaymentService interface {
	// Webhook event handlers
	HandlePaymentIntentSucceeded(paymentIntent *stripe.PaymentIntent) error
	HandlePaymentIntentFailed(paymentIntent *stripe.PaymentIntent) error
	HandlePaymentIntentCanceled(paymentIntent *stripe.PaymentIntent) error
	HandleChargeSucceeded(charge *stripe.Charge) error
	HandleChargeFailed(charge *stripe.Charge) error
	HandleChargeRefunded(charge *stripe.Charge) error
}

type paymentService struct {
	*Service
}

func NewPaymentService(s *Service) PaymentService {
	return &paymentService{
		Service: s,
	}
}

// HandlePaymentIntentSucceeded updates payment records when a payment intent succeeds
func (s *paymentService) HandlePaymentIntentSucceeded(paymentIntent *stripe.PaymentIntent) error {
	s.logger.Infof("Handling payment intent succeeded: %s", paymentIntent.ID)

	// Try to find the payment by payment intent ID
	payment, err := s.store.Payments.GetByStripePaymentIntentID(paymentIntent.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Payment record might not exist yet - this is OK as it might be created later
			s.logger.Infof("Payment record not found for payment intent %s, will be created by other webhook", paymentIntent.ID)
			return nil
		}
		return fmt.Errorf("failed to get payment by payment intent ID: %w", err)
	}

	// Update payment status and details
	if payment.Status != models.PaymentStatusCompleted {
		payment.Status = models.PaymentStatusCompleted
		now := time.Now()
		payment.CompletedAt = &now
		if payment.ProcessedAt == nil {
			payment.ProcessedAt = &now
		}

		// Update payment method type if available
		if paymentIntent.PaymentMethod != nil {
			payment.PaymentMethodType = mapStripePaymentMethodType(string(paymentIntent.PaymentMethod.Type))
		}

		// Get charge ID from latest charge
		if paymentIntent.LatestCharge != nil && paymentIntent.LatestCharge.ID != "" {
			payment.StripeChargeID = paymentIntent.LatestCharge.ID
		}

		// Update metadata with payment intent details
		if err := updatePaymentMetadata(payment, map[string]interface{}{
			"payment_intent_id":     paymentIntent.ID,
			"payment_intent_status": string(paymentIntent.Status),
		}); err != nil {
			s.logger.Errorf("Failed to update payment metadata: %v", err)
		}

		if err := s.store.Payments.Update(payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		s.logger.Infof("Payment %d updated for payment intent %s", payment.ID, paymentIntent.ID)
	}

	return nil
}

// HandlePaymentIntentFailed updates payment records when a payment intent fails
func (s *paymentService) HandlePaymentIntentFailed(paymentIntent *stripe.PaymentIntent) error {
	s.logger.Warnf("Handling payment intent failed: %s", paymentIntent.ID)

	// Try to find the payment by payment intent ID
	payment, err := s.store.Payments.GetByStripePaymentIntentID(paymentIntent.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Payment record not found for failed payment intent %s", paymentIntent.ID)
			return nil
		}
		return fmt.Errorf("failed to get payment by payment intent ID: %w", err)
	}

	// Update payment status with failure information
	payment.Status = models.PaymentStatusFailed
	now := time.Now()
	payment.ProcessedAt = &now

	// Extract failure information from LastPaymentError
	if paymentIntent.LastPaymentError != nil {
		payment.FailureReason = paymentIntent.LastPaymentError.Msg
		payment.FailureCode = string(paymentIntent.LastPaymentError.Code)
	}

	// Update metadata
	metadataUpdates := map[string]interface{}{
		"payment_intent_id":     paymentIntent.ID,
		"payment_intent_status": string(paymentIntent.Status),
	}
	if paymentIntent.LastPaymentError != nil {
		metadataUpdates["failure_message"] = paymentIntent.LastPaymentError.Msg
		metadataUpdates["failure_code"] = string(paymentIntent.LastPaymentError.Code)
	}
	if err := updatePaymentMetadata(payment, metadataUpdates); err != nil {
		s.logger.Errorf("Failed to update payment metadata: %v", err)
	}

	if err := s.store.Payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	s.logger.Infof("Payment %d marked as failed for payment intent %s", payment.ID, paymentIntent.ID)
	return nil
}

// HandlePaymentIntentCanceled updates payment records when a payment intent is canceled
func (s *paymentService) HandlePaymentIntentCanceled(paymentIntent *stripe.PaymentIntent) error {
	s.logger.Infof("Handling payment intent canceled: %s", paymentIntent.ID)

	// Try to find the payment by payment intent ID
	payment, err := s.store.Payments.GetByStripePaymentIntentID(paymentIntent.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Infof("Payment record not found for canceled payment intent %s", paymentIntent.ID)
			return nil
		}
		return fmt.Errorf("failed to get payment by payment intent ID: %w", err)
	}

	// Update payment status to canceled
	payment.Status = models.PaymentStatusCanceled
	now := time.Now()
	payment.ProcessedAt = &now

	// Update metadata
	if err := updatePaymentMetadata(payment, map[string]interface{}{
		"payment_intent_id":     paymentIntent.ID,
		"payment_intent_status": string(paymentIntent.Status),
		"cancellation_reason":   string(paymentIntent.CancellationReason),
	}); err != nil {
		s.logger.Errorf("Failed to update payment metadata: %v", err)
	}

	if err := s.store.Payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	s.logger.Infof("Payment %d marked as canceled for payment intent %s", payment.ID, paymentIntent.ID)
	return nil
}

// HandleChargeSucceeded updates payment records when a charge succeeds
func (s *paymentService) HandleChargeSucceeded(charge *stripe.Charge) error {
	s.logger.Infof("Handling charge succeeded: %s", charge.ID)

	// Try to find payment by charge ID first
	payment, err := s.store.Payments.GetByStripeChargeID(charge.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Try to find by payment intent ID
			if charge.PaymentIntent != nil && charge.PaymentIntent.ID != "" {
				payment, err = s.store.Payments.GetByStripePaymentIntentID(charge.PaymentIntent.ID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						s.logger.Infof("Payment record not found for charge %s", charge.ID)
						return nil
					}
					return fmt.Errorf("failed to get payment by payment intent ID: %w", err)
				}
			} else {
				s.logger.Infof("Payment record not found for charge %s and no payment intent", charge.ID)
				return nil
			}
		} else {
			return fmt.Errorf("failed to get payment by charge ID: %w", err)
		}
	}

	// Update payment with charge details
	payment.StripeChargeID = charge.ID
	if payment.Status != models.PaymentStatusCompleted {
		payment.Status = models.PaymentStatusCompleted
		now := time.Now()
		payment.CompletedAt = &now
		if payment.ProcessedAt == nil {
			payment.ProcessedAt = &now
		}
	}

	// Update payment method information
	if charge.PaymentMethodDetails != nil {
		if charge.PaymentMethodDetails.Card != nil {
			payment.PaymentMethodType = models.PaymentMethodCard
			payment.PaymentMethodLast4 = charge.PaymentMethodDetails.Card.Last4
		} else if charge.PaymentMethodDetails.Type != "" {
			payment.PaymentMethodType = mapStripePaymentMethodType(string(charge.PaymentMethodDetails.Type))
		}
	}

	// Update metadata
	if err := updatePaymentMetadata(payment, map[string]interface{}{
		"charge_id":       charge.ID,
		"charge_status":   string(charge.Status),
		"amount_captured": charge.AmountCaptured,
	}); err != nil {
		s.logger.Errorf("Failed to update payment metadata: %v", err)
	}

	if err := s.store.Payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	s.logger.Infof("Payment %d updated for charge %s", payment.ID, charge.ID)
	return nil
}

// HandleChargeFailed updates payment records when a charge fails
func (s *paymentService) HandleChargeFailed(charge *stripe.Charge) error {
	s.logger.Warnf("Handling charge failed: %s", charge.ID)

	// Try to find payment by charge ID first
	payment, err := s.store.Payments.GetByStripeChargeID(charge.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Try to find by payment intent ID
			if charge.PaymentIntent != nil && charge.PaymentIntent.ID != "" {
				payment, err = s.store.Payments.GetByStripePaymentIntentID(charge.PaymentIntent.ID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						s.logger.Warnf("Payment record not found for failed charge %s", charge.ID)
						return nil
					}
					return fmt.Errorf("failed to get payment by payment intent ID: %w", err)
				}
			} else {
				s.logger.Warnf("Payment record not found for failed charge %s and no payment intent", charge.ID)
				return nil
			}
		} else {
			return fmt.Errorf("failed to get payment by charge ID: %w", err)
		}
	}

	// Update payment with failure information
	payment.StripeChargeID = charge.ID
	payment.Status = models.PaymentStatusFailed
	now := time.Now()
	payment.ProcessedAt = &now

	// Extract failure information
	if charge.FailureMessage != "" {
		payment.FailureReason = charge.FailureMessage
	}
	if charge.FailureCode != "" {
		payment.FailureCode = charge.FailureCode
	}

	// Update metadata
	if err := updatePaymentMetadata(payment, map[string]interface{}{
		"charge_id":       charge.ID,
		"charge_status":   string(charge.Status),
		"failure_message": charge.FailureMessage,
		"failure_code":    charge.FailureCode,
	}); err != nil {
		s.logger.Errorf("Failed to update payment metadata: %v", err)
	}

	if err := s.store.Payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	s.logger.Infof("Payment %d marked as failed for charge %s", payment.ID, charge.ID)
	return nil
}

// HandleChargeRefunded updates payment records when a charge is refunded
func (s *paymentService) HandleChargeRefunded(charge *stripe.Charge) error {
	s.logger.Infof("Handling charge refunded: %s", charge.ID)

	// Try to find payment by charge ID first
	payment, err := s.store.Payments.GetByStripeChargeID(charge.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Try to find by payment intent ID
			if charge.PaymentIntent != nil && charge.PaymentIntent.ID != "" {
				payment, err = s.store.Payments.GetByStripePaymentIntentID(charge.PaymentIntent.ID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						s.logger.Warnf("Payment record not found for refunded charge %s", charge.ID)
						return nil
					}
					return fmt.Errorf("failed to get payment by payment intent ID: %w", err)
				}
			} else {
				s.logger.Warnf("Payment record not found for refunded charge %s and no payment intent", charge.ID)
				return nil
			}
		} else {
			return fmt.Errorf("failed to get payment by charge ID: %w", err)
		}
	}

	// Update payment with refund information
	payment.StripeChargeID = charge.ID
	payment.Status = models.PaymentStatusRefunded
	now := time.Now()
	payment.RefundedAt = &now

	// Calculate refund amount - charge.AmountRefunded is in cents
	payment.RefundAmount = charge.AmountRefunded

	// Update metadata with refund details
	metadataUpdates := map[string]interface{}{
		"charge_id":       charge.ID,
		"charge_status":   string(charge.Status),
		"amount_refunded": charge.AmountRefunded,
		"refunded_at":     now.Format(time.RFC3339),
	}

	// Add refund details if available
	if charge.Refunds != nil && len(charge.Refunds.Data) > 0 {
		refundDetails := make([]map[string]interface{}, 0)
		for _, refund := range charge.Refunds.Data {
			refundDetails = append(refundDetails, map[string]interface{}{
				"id":     refund.ID,
				"amount": refund.Amount,
				"reason": string(refund.Reason),
				"status": string(refund.Status),
			})
		}
		metadataUpdates["refunds"] = refundDetails
	}

	if err := updatePaymentMetadata(payment, metadataUpdates); err != nil {
		s.logger.Errorf("Failed to update payment metadata: %v", err)
	}

	if err := s.store.Payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	s.logger.Infof("Payment %d marked as refunded for charge %s, amount: %d", payment.ID, charge.ID, charge.AmountRefunded)
	return nil
}

// Helper function to update payment metadata with new information
func updatePaymentMetadata(payment *models.Payment, updates map[string]interface{}) error {
	metadata := make(map[string]interface{})

	// Parse existing metadata if it exists
	if payment.Metadata != "" && payment.Metadata != "{}" {
		if err := json.Unmarshal([]byte(payment.Metadata), &metadata); err != nil {
			return fmt.Errorf("failed to unmarshal existing metadata: %w", err)
		}
	}

	// Merge updates into metadata
	for key, value := range updates {
		metadata[key] = value
	}

	// Marshal back to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	payment.Metadata = string(metadataJSON)
	return nil
}

// Helper function to map Stripe payment method types to our internal types
func mapStripePaymentMethodType(stripeType string) models.PaymentMethodType {
	if stripeType == "card" {
		return models.PaymentMethodCard
	}

	// Check if it's a crypto payment method using the stripe package helper
	if stripeClient.IsCryptoPaymentMethod(stripeType) {
		return models.PaymentMethodCrypto
	}

	return models.PaymentMethodUnknown
}
