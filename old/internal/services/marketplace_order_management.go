package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/email"
	"github.com/nicolailuther/butter/pkg/email/templates"
)

type MarketplaceOrderManagementService interface {
	// Deliverable operations
	SubmitDeliverable(data *dto.CreateMarketplaceDeliverable, sellerID uint) (*models.MarketplaceDeliverable, error)
	ReviewDeliverable(deliverableID uint, data *dto.ReviewMarketplaceDeliverable, buyerID uint) (*models.MarketplaceDeliverable, error)
	GetDeliverablesByOrder(orderID uint) ([]*models.MarketplaceDeliverable, error)
	GetDeliverableByID(deliverableID uint) (*models.MarketplaceDeliverable, error)

	// Revision operations
	RequestRevision(orderID uint, data *dto.CreateRevisionRequest, buyerID uint) (*models.MarketplaceRevisionRequest, error)
	RespondToRevision(revisionID uint, data *dto.RespondToRevision, sellerID uint) (*models.MarketplaceRevisionRequest, error)
	GetRevisionsByOrder(orderID uint) ([]*models.MarketplaceRevisionRequest, error)

	// Dispute operations
	OpenDispute(data *dto.CreateDispute, buyerID uint) (*models.MarketplaceDispute, error)
	ResolveDispute(disputeID uint, data *dto.ResolveDispute, adminID uint) (*models.MarketplaceDispute, error)
	GetDisputeByID(disputeID uint) (*models.MarketplaceDispute, error)
	GetAllDisputes() ([]*models.MarketplaceDispute, error)
	GetDisputeByOrder(orderID uint) (*models.MarketplaceDispute, error)

	// Order management operations
	StartOrder(orderID uint, sellerID uint) (*models.MarketplaceOrder, error)
	CompleteOrder(orderID uint, buyerID uint) (*models.MarketplaceOrder, error)
	CancelOrder(orderID uint, data *dto.CancelOrder, userID uint) (*models.MarketplaceOrder, error)
	ExtendDeadline(orderID uint, data *dto.ExtendDeadline, sellerID uint) (*models.MarketplaceOrder, error)

	// Timeline operations
	GetOrderTimeline(orderID uint) ([]*models.MarketplaceOrderTimeline, error)
	AddTimelineEvent(orderID uint, eventType enums.OrderTimelineEventType, description string, actorID uint, metadata map[string]interface{}) error
}

type marketplaceOrderManagementServiceImpl struct {
	*Service
	notificationService NotificationService
	fileService         FileService
	marketplaceService  MarketplaceService
	emailClient         EmailClient
}

func NewMarketplaceOrderManagementService(
	container *Service,
	notificationService NotificationService,
	fileService FileService,
	marketplaceService MarketplaceService,
	emailClient EmailClient,
) MarketplaceOrderManagementService {
	return &marketplaceOrderManagementServiceImpl{
		container,
		notificationService,
		fileService,
		marketplaceService,
		emailClient,
	}
}

// SubmitDeliverable allows a seller to submit work deliverable for an order
func (s *marketplaceOrderManagementServiceImpl) SubmitDeliverable(data *dto.CreateMarketplaceDeliverable, sellerID uint) (*models.MarketplaceDeliverable, error) {
	// Validate order
	order, err := s.store.MarketplaceOrders.GetByID(data.OrderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify seller owns the service
	if order.Service.UserID != sellerID {
		return nil, errors.New("only the service owner can submit deliverables")
	}

	// Check order status
	if order.Status != enums.MarketplaceOrderStatusInProgress && order.Status != enums.MarketplaceOrderStatusInRevision {
		return nil, errors.New("order must be in progress or in revision to submit deliverable")
	}

	// Check for existing pending deliverable
	pending, _ := s.store.MarketplaceDeliverables.GetPendingByOrder(data.OrderID)
	if pending != nil {
		return nil, errors.New("there is already a pending deliverable for this order")
	}

	// Validate description length
	if len(data.Description) < domain.DeliveryDescriptionMinLength {
		return nil, fmt.Errorf("description must be at least %d characters", domain.DeliveryDescriptionMinLength)
	}

	// Create deliverable
	deliverable := &models.MarketplaceDeliverable{
		OrderID:      data.OrderID,
		SellerID:     sellerID,
		Description:  data.Description,
		DeliveryNote: data.DeliveryNote,
		Status:       enums.DeliverableStatusPending,
	}

	if err := s.store.MarketplaceDeliverables.Create(deliverable); err != nil {
		return nil, err
	}

	// Update order status and set auto-completion date
	order.Status = enums.MarketplaceOrderStatusDelivered
	order.TotalDeliveries++
	order.AutoCompletionDate = time.Now().AddDate(0, 0, domain.AutoCompletionDays)
	s.store.MarketplaceOrders.Update(order)

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventDeliverySubmitted,
		"Seller submitted deliverable", sellerID, map[string]interface{}{
			"deliverable_id": deliverable.ID,
		})

	// Notify buyer with rich email template
	subject, htmlBody, textBody := templates.DeliverySubmitted(templates.MarketplaceEmailData{
		OrderID:      order.ID,
		ServiceTitle: order.Service.Title,
		SellerName:   order.Service.User.Name,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.BuyerID,
		"Delivery Submitted",
		fmt.Sprintf("The seller has submitted work for your order #%d", order.ID),
		subject,
		htmlBody,
		textBody,
	)

	return deliverable, nil
}

// ReviewDeliverable allows a buyer to review and accept/reject a deliverable
func (s *marketplaceOrderManagementServiceImpl) ReviewDeliverable(deliverableID uint, data *dto.ReviewMarketplaceDeliverable, buyerID uint) (*models.MarketplaceDeliverable, error) {
	// Get deliverable
	deliverable, err := s.store.MarketplaceDeliverables.GetByID(deliverableID)
	if err != nil {
		return nil, errors.New("deliverable not found")
	}

	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(deliverable.OrderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify buyer owns the order
	if order.BuyerID != buyerID {
		return nil, errors.New("only the order buyer can review deliverables")
	}

	// Check deliverable status
	if deliverable.Status != enums.DeliverableStatusPending {
		return nil, errors.New("deliverable has already been reviewed")
	}

	// Update deliverable
	deliverable.ReviewNotes = data.ReviewNotes
	deliverable.ReviewedAt = time.Now()

	if data.Accept {
		deliverable.Status = enums.DeliverableStatusAccepted

		// Complete the order
		order.Status = enums.MarketplaceOrderStatusCompleted
		order.CompletedAt = time.Now()
		s.store.MarketplaceOrders.Update(order)

		// Add timeline event
		s.AddTimelineEvent(order.ID, enums.OrderTimelineEventDeliveryReviewed,
			"Buyer accepted deliverable", buyerID, map[string]interface{}{
				"deliverable_id": deliverable.ID,
				"accepted":       true,
			})

		s.AddTimelineEvent(order.ID, enums.OrderTimelineEventCompleted,
			"Order completed successfully", buyerID, nil)

		// Notify seller with rich email template
		subject, htmlBody, textBody := templates.DeliveryAccepted(templates.MarketplaceEmailData{
			OrderID:      order.ID,
			ServiceTitle: order.Service.Title,
			BuyerName:    order.Buyer.Name,
		})
		s.sendNotificationAndEmailWithTemplate(
			order.Service.UserID,
			"Delivery Accepted",
			fmt.Sprintf("Your deliverable for order #%d has been accepted!", order.ID),
			subject,
			htmlBody,
			textBody,
		)
	} else {
		deliverable.Status = enums.DeliverableStatusRejected

		// Update order status back to in_progress and clear auto-completion
		order.Status = enums.MarketplaceOrderStatusInProgress
		order.AutoCompletionDate = time.Time{} // Clear auto-completion date
		s.store.MarketplaceOrders.Update(order)

		// Add timeline event
		s.AddTimelineEvent(order.ID, enums.OrderTimelineEventDeliveryReviewed,
			"Buyer rejected deliverable", buyerID, map[string]interface{}{
				"deliverable_id": deliverable.ID,
				"accepted":       false,
			})

		// Notify seller with rich email template
		feedbackInfo := ""
		if data.ReviewNotes != "" {
			feedbackInfo = fmt.Sprintf("<strong>Buyer's Feedback:</strong><br>%s", data.ReviewNotes)
		}
		subject, htmlBody, textBody := templates.DeliveryRejected(templates.MarketplaceEmailData{
			OrderID:        order.ID,
			ServiceTitle:   order.Service.Title,
			BuyerName:      order.Buyer.Name,
			AdditionalInfo: feedbackInfo,
		})
		s.sendNotificationAndEmailWithTemplate(
			order.Service.UserID,
			"Delivery Rejected",
			fmt.Sprintf("Your deliverable for order #%d was rejected. Please review the feedback.", order.ID),
			subject,
			htmlBody,
			textBody,
		)
	}

	if err := s.store.MarketplaceDeliverables.Update(deliverable); err != nil {
		return nil, err
	}

	return deliverable, nil
}

// GetDeliverablesByOrder retrieves all deliverables for an order
func (s *marketplaceOrderManagementServiceImpl) GetDeliverablesByOrder(orderID uint) ([]*models.MarketplaceDeliverable, error) {
	if orderID == 0 {
		return nil, errors.New("order_id is required")
	}

	deliverables, err := s.store.MarketplaceDeliverables.GetAllByOrder(orderID)
	if err != nil {
		return nil, err
	}

	return deliverables, nil
}

// GetDeliverableByID retrieves a specific deliverable
func (s *marketplaceOrderManagementServiceImpl) GetDeliverableByID(deliverableID uint) (*models.MarketplaceDeliverable, error) {
	if deliverableID == 0 {
		return nil, errors.New("deliverable_id is required")
	}

	return s.store.MarketplaceDeliverables.GetByID(deliverableID)
}

// RequestRevision allows a buyer to request changes to a deliverable
func (s *marketplaceOrderManagementServiceImpl) RequestRevision(orderID uint, data *dto.CreateRevisionRequest, buyerID uint) (*models.MarketplaceRevisionRequest, error) {
	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify buyer owns the order
	if order.BuyerID != buyerID {
		return nil, errors.New("only the order buyer can request revisions")
	}

	// Get deliverable
	deliverable, err := s.store.MarketplaceDeliverables.GetByID(data.DeliverableID)
	if err != nil {
		return nil, errors.New("deliverable not found")
	}

	// Verify deliverable belongs to order
	if deliverable.OrderID != orderID {
		return nil, errors.New("deliverable does not belong to this order")
	}

	// Check if revisions are available
	revisionsRemaining := order.ServicePackage.IncludedRevisions - order.RevisionsUsed
	if revisionsRemaining <= 0 {
		return nil, errors.New("no revisions remaining for this order")
	}

	// Validate reason length
	if len(data.Reason) < domain.RevisionReasonMinLength {
		return nil, fmt.Errorf("revision reason must be at least %d characters", domain.RevisionReasonMinLength)
	}

	// Create revision request
	revision := &models.MarketplaceRevisionRequest{
		OrderID:        orderID,
		DeliverableID:  data.DeliverableID,
		RevisionNumber: order.RevisionsUsed + 1,
		Reason:         data.Reason,
		Status:         enums.RevisionRequestStatusPending,
	}

	if err := s.store.MarketplaceRevisionRequests.Create(revision); err != nil {
		return nil, err
	}

	// Update order - clear auto-completion date since order is now in revision
	order.Status = enums.MarketplaceOrderStatusInRevision
	order.RevisionsUsed++
	order.AutoCompletionDate = time.Time{} // Clear auto-completion date
	s.store.MarketplaceOrders.Update(order)

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventRevisionRequested,
		fmt.Sprintf("Buyer requested revision #%d", revision.RevisionNumber), buyerID, map[string]interface{}{
			"revision_id":     revision.ID,
			"revision_number": revision.RevisionNumber,
		})

	// Notify seller with rich email template
	revisionInfo := fmt.Sprintf("<strong>Revision #%d:</strong><br>%s", revision.RevisionNumber, html.EscapeString(data.Reason))
	subject, htmlBody, textBody := templates.RevisionRequested(templates.MarketplaceEmailData{
		OrderID:        order.ID,
		ServiceTitle:   order.Service.Title,
		BuyerName:      order.Buyer.Name,
		AdditionalInfo: revisionInfo,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.Service.UserID,
		"Revision Requested",
		fmt.Sprintf("Buyer requested changes for order #%d (Revision #%d)", order.ID, revision.RevisionNumber),
		subject,
		htmlBody,
		textBody,
	)

	return revision, nil
}

// RespondToRevision allows a seller to respond to a revision request
func (s *marketplaceOrderManagementServiceImpl) RespondToRevision(revisionID uint, data *dto.RespondToRevision, sellerID uint) (*models.MarketplaceRevisionRequest, error) {
	// Get revision
	revision, err := s.store.MarketplaceRevisionRequests.GetByID(revisionID)
	if err != nil {
		return nil, errors.New("revision request not found")
	}

	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(revision.OrderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify seller owns the service
	if order.Service.UserID != sellerID {
		return nil, errors.New("only the service owner can respond to revisions")
	}

	// Check revision status
	if revision.Status != enums.RevisionRequestStatusPending {
		return nil, errors.New("revision has already been responded to")
	}

	// Update revision
	revision.SellerResponse = data.SellerResponse
	revision.RespondedAt = time.Now()

	if data.Accept {
		revision.Status = enums.RevisionRequestStatusAccepted

		// Notify buyer with rich email template
		subject, htmlBody, textBody := templates.RevisionAccepted(templates.MarketplaceEmailData{
			OrderID:      order.ID,
			ServiceTitle: order.Service.Title,
			SellerName:   order.Service.User.Name,
		})
		s.sendNotificationAndEmailWithTemplate(
			order.BuyerID,
			"Revision Accepted",
			fmt.Sprintf("Seller accepted your revision request for order #%d", order.ID),
			subject,
			htmlBody,
			textBody,
		)
	}

	if err := s.store.MarketplaceRevisionRequests.Update(revision); err != nil {
		return nil, err
	}

	return revision, nil
}

// GetRevisionsByOrder retrieves all revisions for an order
func (s *marketplaceOrderManagementServiceImpl) GetRevisionsByOrder(orderID uint) ([]*models.MarketplaceRevisionRequest, error) {
	if orderID == 0 {
		return nil, errors.New("order_id is required")
	}

	return s.store.MarketplaceRevisionRequests.GetAllByOrder(orderID)
}

// OpenDispute allows a buyer to open a dispute for an order
func (s *marketplaceOrderManagementServiceImpl) OpenDispute(data *dto.CreateDispute, buyerID uint) (*models.MarketplaceDispute, error) {
	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(data.OrderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify buyer owns the order
	if order.BuyerID != buyerID {
		return nil, errors.New("only the order buyer can open disputes")
	}

	// Check order status
	if order.Status == enums.MarketplaceOrderStatusCompleted ||
		order.Status == enums.MarketplaceOrderStatusCancelled ||
		order.Status == enums.MarketplaceOrderStatusRefunded {
		return nil, errors.New("cannot open dispute for completed, cancelled or refunded orders")
	}

	// Check for existing dispute
	existingDispute, _ := s.store.MarketplaceDisputes.GetByOrder(data.OrderID)
	if existingDispute != nil {
		return nil, errors.New("a dispute already exists for this order")
	}

	// Validate description length
	if len(data.Description) < domain.DisputeDescriptionMinLength {
		return nil, fmt.Errorf("dispute description must be at least %d characters", domain.DisputeDescriptionMinLength)
	}

	// Validate reason category
	if !data.ReasonCategory.IsValid() {
		return nil, errors.New("invalid dispute reason category")
	}

	// Create dispute
	dispute := &models.MarketplaceDispute{
		OrderID:        data.OrderID,
		OpenedBy:       buyerID,
		ReasonCategory: data.ReasonCategory,
		Description:    data.Description,
		Status:         enums.DisputeStatusOpen,
	}

	if err := s.store.MarketplaceDisputes.Create(dispute); err != nil {
		return nil, err
	}

	// Update order status
	order.Status = enums.MarketplaceOrderStatusDisputed
	s.store.MarketplaceOrders.Update(order)

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventDisputeOpened,
		"Buyer opened a dispute", buyerID, map[string]interface{}{
			"dispute_id":      dispute.ID,
			"reason_category": data.ReasonCategory,
		})

	// Notify seller with rich email template
	disputeInfo := fmt.Sprintf("<strong>Reason:</strong> %s<br><strong>Description:</strong><br>%s",
		html.EscapeString(string(data.ReasonCategory)),
		html.EscapeString(data.Description))
	subject, htmlBody, textBody := templates.DisputeOpened(templates.MarketplaceEmailData{
		OrderID:        order.ID,
		ServiceTitle:   order.Service.Title,
		AdditionalInfo: disputeInfo,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.Service.UserID,
		"Dispute Opened",
		fmt.Sprintf("A dispute has been opened for order #%d", order.ID),
		subject,
		htmlBody,
		textBody,
	)

	return dispute, nil
}

// ResolveDispute allows an admin to resolve a dispute
func (s *marketplaceOrderManagementServiceImpl) ResolveDispute(disputeID uint, data *dto.ResolveDispute, adminID uint) (*models.MarketplaceDispute, error) {
	// Get dispute
	dispute, err := s.store.MarketplaceDisputes.GetByID(disputeID)
	if err != nil {
		return nil, errors.New("dispute not found")
	}

	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(dispute.OrderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Validate resolution
	validResolutions := map[enums.DisputeResolution]bool{
		enums.DisputeResolutionBuyerFavor:  true,
		enums.DisputeResolutionSellerFavor: true,
		enums.DisputeResolutionSplit:       true,
	}
	if !validResolutions[data.Resolution] {
		return nil, errors.New("invalid dispute resolution")
	}

	// Update dispute
	dispute.Status = enums.DisputeStatusResolved
	dispute.Resolution = data.Resolution
	dispute.ResolutionNotes = data.ResolutionNotes
	dispute.RefundAmount = data.RefundAmount
	dispute.ResolvedBy = adminID
	dispute.ResolvedAt = time.Now()

	if err := s.store.MarketplaceDisputes.Update(dispute); err != nil {
		return nil, err
	}

	// Update order based on resolution
	switch data.Resolution {
	case enums.DisputeResolutionBuyerFavor:
		order.Status = enums.MarketplaceOrderStatusRefunded
	case enums.DisputeResolutionSellerFavor:
		order.Status = enums.MarketplaceOrderStatusCompleted
		order.CompletedAt = time.Now()
	case enums.DisputeResolutionSplit:
		order.Status = enums.MarketplaceOrderStatusCompleted
		order.CompletedAt = time.Now()
	}
	s.store.MarketplaceOrders.Update(order)

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventDisputeResolved,
		fmt.Sprintf("Dispute resolved in %s", data.Resolution), adminID, map[string]interface{}{
			"dispute_id":    dispute.ID,
			"resolution":    data.Resolution,
			"refund_amount": data.RefundAmount,
		})

	// Notify both parties with rich email template
	resolutionInfo := fmt.Sprintf("<strong>Resolution:</strong> %s", html.EscapeString(string(data.Resolution)))
	if data.ResolutionNotes != "" {
		resolutionInfo += fmt.Sprintf("<br><strong>Notes:</strong><br>%s", html.EscapeString(data.ResolutionNotes))
	}
	if data.RefundAmount > 0 {
		resolutionInfo += fmt.Sprintf("<br><strong>Refund Amount:</strong> $%.2f", data.RefundAmount)
	}

	buyerSubject, buyerHtmlBody, buyerTextBody := templates.DisputeResolved(templates.MarketplaceEmailData{
		OrderID:        order.ID,
		ServiceTitle:   order.Service.Title,
		AdditionalInfo: resolutionInfo,
	}, true)

	s.sendNotificationAndEmailWithTemplate(
		order.BuyerID,
		"Dispute Resolved",
		fmt.Sprintf("The dispute for order #%d has been resolved", order.ID),
		buyerSubject,
		buyerHtmlBody,
		buyerTextBody,
	)

	sellerSubject, sellerHtmlBody, sellerTextBody := templates.DisputeResolved(templates.MarketplaceEmailData{
		OrderID:        order.ID,
		ServiceTitle:   order.Service.Title,
		AdditionalInfo: resolutionInfo,
	}, false)

	s.sendNotificationAndEmailWithTemplate(
		order.Service.UserID,
		"Dispute Resolved",
		fmt.Sprintf("The dispute for order #%d has been resolved", order.ID),
		sellerSubject,
		sellerHtmlBody,
		sellerTextBody,
	)

	return dispute, nil
}

// GetDisputeByID retrieves a specific dispute
func (s *marketplaceOrderManagementServiceImpl) GetDisputeByID(disputeID uint) (*models.MarketplaceDispute, error) {
	if disputeID == 0 {
		return nil, errors.New("dispute_id is required")
	}

	return s.store.MarketplaceDisputes.GetByID(disputeID)
}

// GetAllDisputes retrieves all disputes (admin only)
func (s *marketplaceOrderManagementServiceImpl) GetAllDisputes() ([]*models.MarketplaceDispute, error) {
	return s.store.MarketplaceDisputes.GetAll()
}

// GetDisputeByOrder retrieves the dispute for a specific order
func (s *marketplaceOrderManagementServiceImpl) GetDisputeByOrder(orderID uint) (*models.MarketplaceDispute, error) {
	if orderID == 0 {
		return nil, errors.New("order_id is required")
	}

	return s.store.MarketplaceDisputes.GetByOrder(orderID)
}

// StartOrder allows seller to start working on a paid order
func (s *marketplaceOrderManagementServiceImpl) StartOrder(orderID uint, sellerID uint) (*models.MarketplaceOrder, error) {
	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify seller owns the service
	if order.Service.UserID != sellerID {
		return nil, errors.New("only the service owner can start this order")
	}

	// Check order status - can start from 'paid' or 'waiting_info'
	if order.Status != enums.MarketplaceOrderStatusPaid && order.Status != enums.MarketplaceOrderStatusWaitingInfo {
		return nil, errors.New("order must be in paid or waiting_info status to start")
	}

	// Calculate deadline based on package duration
	order.StartDate = time.Now()
	if order.ServicePackage != nil && order.ServicePackage.DurationDays > 0 {
		order.DueDate = time.Now().Add(time.Duration(order.ServicePackage.DurationDays) * 24 * time.Hour)
	}

	// Initialize revision counts from package
	if order.ServicePackage != nil {
		order.TotalRevisions = order.ServicePackage.IncludedRevisions
	} else {
		order.TotalRevisions = domain.MaxRevisionsDefault
	}

	// Update order status to in_progress
	order.Status = enums.MarketplaceOrderStatusInProgress

	if err := s.store.MarketplaceOrders.Update(order); err != nil {
		return nil, err
	}

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventStatusChanged,
		"Seller started working on order", sellerID, map[string]interface{}{
			"from_status": enums.MarketplaceOrderStatusPaid,
			"to_status":   enums.MarketplaceOrderStatusInProgress,
			"due_date":    order.DueDate,
		})

	// Notify buyer with rich email template
	subject, htmlBody, textBody := templates.OrderStarted(templates.MarketplaceEmailData{
		OrderID:      order.ID,
		ServiceTitle: order.Service.Title,
		SellerName:   order.Service.User.Name,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.BuyerID,
		"Order Started",
		fmt.Sprintf("Seller has started working on your order #%d", order.ID),
		subject,
		htmlBody,
		textBody,
	)

	return order, nil
}

// CompleteOrder allows a buyer to manually complete an order
func (s *marketplaceOrderManagementServiceImpl) CompleteOrder(orderID uint, buyerID uint) (*models.MarketplaceOrder, error) {
	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify buyer owns the order
	if order.BuyerID != buyerID {
		return nil, errors.New("only the order buyer can complete orders")
	}

	// Check order status
	if order.Status != enums.MarketplaceOrderStatusDelivered {
		return nil, errors.New("order must have a delivered status to be completed")
	}

	// Update order
	order.Status = enums.MarketplaceOrderStatusCompleted
	order.CompletedAt = time.Now()

	if err := s.store.MarketplaceOrders.Update(order); err != nil {
		return nil, err
	}

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventCompleted,
		"Order completed by buyer", buyerID, nil)

	// Notify seller with rich email template
	subject, htmlBody, textBody := templates.OrderCompleted(templates.MarketplaceEmailData{
		OrderID:      order.ID,
		ServiceTitle: order.Service.Title,
		BuyerName:    order.Buyer.Name,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.Service.UserID,
		"Order Completed",
		fmt.Sprintf("Order #%d has been completed!", order.ID),
		subject,
		htmlBody,
		textBody,
	)

	return order, nil
}

// CancelOrder allows cancelling an order under specific conditions
func (s *marketplaceOrderManagementServiceImpl) CancelOrder(orderID uint, data *dto.CancelOrder, userID uint) (*models.MarketplaceOrder, error) {
	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Check if user is buyer or seller
	isBuyer := order.BuyerID == userID
	isSeller := order.Service.UserID == userID

	if !isBuyer && !isSeller {
		return nil, errors.New("only buyer or seller can cancel orders")
	}

	// Reason is optional, no minimum length validation
	// Trim whitespace from reason for consistent handling throughout the function
	// Empty or whitespace-only reasons are treated as empty strings and omitted from messages
	trimmedReason := strings.TrimSpace(data.Reason)

	// Check order status - cannot cancel delivered, disputed, completed, etc.
	invalidStatuses := []enums.MarketplaceOrderStatus{
		enums.MarketplaceOrderStatusDelivered,
		enums.MarketplaceOrderStatusInRevision,
		enums.MarketplaceOrderStatusDisputed,
		enums.MarketplaceOrderStatusCompleted,
		enums.MarketplaceOrderStatusCancelled,
		enums.MarketplaceOrderStatusRefunded,
		enums.MarketplaceOrderStatusAutoCompleted,
	}

	for _, status := range invalidStatuses {
		if order.Status == status {
			return nil, fmt.Errorf("cannot cancel order in %s status", status)
		}
	}

	// Determine if refund is needed
	requiresRefund := order.Status == enums.MarketplaceOrderStatusPaid ||
		order.Status == enums.MarketplaceOrderStatusWaitingInfo ||
		order.Status == enums.MarketplaceOrderStatusInProgress

	// Process refund if payment was made
	if requiresRefund && order.PaymentID > 0 {
		// Attempt to refund the payment
		refundReason := "Order cancelled"
		if trimmedReason != "" {
			refundReason = fmt.Sprintf("Order cancelled: %s", trimmedReason)
		}
		err := s.marketplaceService.CreateOrderRefund(order.ID, 0, refundReason)
		if err != nil {
			// Mark as cancelled but note that refund needs manual processing
			// Note: We use 'cancelled' status here because:
			// - The order is cancelled from business perspective (work won't proceed)
			// - Refund status is tracked separately in timeline metadata for admin
			// - This avoids needing a new 'cancelled_pending_refund' status
			order.Status = enums.MarketplaceOrderStatusCancelled
			s.AddTimelineEvent(order.ID, enums.OrderTimelineEventCancelled,
				"Order cancelled successfully - refund pending manual processing", userID, map[string]interface{}{
					"reason":        trimmedReason,
					"refund_error":  err.Error(),
					"refund_status": "pending_manual",
				})
		} else {
			// Update status to refunded if refund succeeded
			order.Status = enums.MarketplaceOrderStatusRefunded
			s.AddTimelineEvent(order.ID, enums.OrderTimelineEventCancelled,
				"Order cancelled and refunded successfully", userID, map[string]interface{}{
					"reason":        trimmedReason,
					"refund_status": "successful",
				})
		}
	} else {
		// No payment made, just cancel
		order.Status = enums.MarketplaceOrderStatusCancelled
		timelineMessage := "Order cancelled successfully"
		if trimmedReason != "" {
			timelineMessage = fmt.Sprintf("Order cancelled successfully: %s", trimmedReason)
		}
		s.AddTimelineEvent(order.ID, enums.OrderTimelineEventCancelled,
			timelineMessage, userID, map[string]interface{}{
				"reason": trimmedReason,
			})
	}

	if err := s.store.MarketplaceOrders.Update(order); err != nil {
		return nil, err
	}

	// Notify the other party with rich email template
	var cancelInfo string
	if trimmedReason != "" {
		cancelInfo = fmt.Sprintf("<strong>Reason:</strong><br>%s", html.EscapeString(trimmedReason))
	}
	if isBuyer {
		if cancelInfo != "" {
			cancelInfo = fmt.Sprintf("<strong>Cancelled by:</strong> Buyer<br>%s", cancelInfo)
		} else {
			cancelInfo = "<strong>Cancelled by:</strong> Buyer"
		}
		subject, htmlBody, textBody := templates.OrderCancelled(templates.MarketplaceEmailData{
			OrderID:        order.ID,
			ServiceTitle:   order.Service.Title,
			AdditionalInfo: cancelInfo,
		}, false)
		s.sendNotificationAndEmailWithTemplate(
			order.Service.UserID,
			"Order Cancelled",
			fmt.Sprintf("Order #%d has been cancelled by the buyer", order.ID),
			subject,
			htmlBody,
			textBody,
		)
	} else {
		if cancelInfo != "" {
			cancelInfo = fmt.Sprintf("<strong>Cancelled by:</strong> Seller<br>%s", cancelInfo)
		} else {
			cancelInfo = "<strong>Cancelled by:</strong> Seller"
		}
		subject, htmlBody, textBody := templates.OrderCancelled(templates.MarketplaceEmailData{
			OrderID:        order.ID,
			ServiceTitle:   order.Service.Title,
			AdditionalInfo: cancelInfo,
		}, true)
		s.sendNotificationAndEmailWithTemplate(
			order.BuyerID,
			"Order Cancelled",
			fmt.Sprintf("Order #%d has been cancelled by the seller", order.ID),
			subject,
			htmlBody,
			textBody,
		)
	}

	return order, nil
}

// ExtendDeadline allows a seller to request deadline extension
func (s *marketplaceOrderManagementServiceImpl) ExtendDeadline(orderID uint, data *dto.ExtendDeadline, sellerID uint) (*models.MarketplaceOrder, error) {
	// Get order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify seller owns the service
	if order.Service.UserID != sellerID {
		return nil, errors.New("only the service owner can request deadline extensions")
	}

	// Validate days
	if data.AdditionalDays < 1 || data.AdditionalDays > domain.MaxDeadlineExtensionDays {
		return nil, fmt.Errorf("additional days must be between 1 and %d", domain.MaxDeadlineExtensionDays)
	}

	// Update order
	if !order.DueDate.IsZero() {
		order.DueDate = order.DueDate.AddDate(0, 0, data.AdditionalDays)
	}
	order.DaysExtended += data.AdditionalDays

	if err := s.store.MarketplaceOrders.Update(order); err != nil {
		return nil, err
	}

	// Add timeline event
	s.AddTimelineEvent(order.ID, enums.OrderTimelineEventDeadlineExtended,
		fmt.Sprintf("Deadline extended by %d days: %s", data.AdditionalDays, data.Reason), sellerID, map[string]interface{}{
			"additional_days": data.AdditionalDays,
			"reason":          data.Reason,
		})

	// Notify buyer with rich email template
	extensionInfo := fmt.Sprintf("<strong>Extended by:</strong> %d days<br><strong>Reason:</strong><br>%s",
		data.AdditionalDays,
		html.EscapeString(data.Reason))
	subject, htmlBody, textBody := templates.DeadlineExtended(templates.MarketplaceEmailData{
		OrderID:        order.ID,
		ServiceTitle:   order.Service.Title,
		DueDate:        &order.DueDate,
		AdditionalInfo: extensionInfo,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.BuyerID,
		"Deadline Extended",
		fmt.Sprintf("The deadline for order #%d has been extended by %d days", order.ID, data.AdditionalDays),
		subject,
		htmlBody,
		textBody,
	)

	return order, nil
}

// GetOrderTimeline retrieves all timeline events for an order
func (s *marketplaceOrderManagementServiceImpl) GetOrderTimeline(orderID uint) ([]*models.MarketplaceOrderTimeline, error) {
	if orderID == 0 {
		return nil, errors.New("order_id is required")
	}

	return s.store.MarketplaceOrderTimelines.GetAllByOrder(orderID)
}

// AddTimelineEvent adds a new event to the order timeline
func (s *marketplaceOrderManagementServiceImpl) AddTimelineEvent(orderID uint, eventType enums.OrderTimelineEventType, description string, actorID uint, metadata map[string]interface{}) error {
	if orderID == 0 {
		return errors.New("order_id is required")
	}

	var metadataJSON string
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err == nil {
			metadataJSON = string(jsonBytes)
		}
	}

	timeline := &models.MarketplaceOrderTimeline{
		OrderID:     orderID,
		EventType:   eventType,
		Description: description,
		ActorID:     actorID,
		Metadata:    metadataJSON,
	}

	return s.store.MarketplaceOrderTimelines.Create(timeline)
}

// sendNotificationAndEmail sends both a notification and an email to a user
func (s *marketplaceOrderManagementServiceImpl) sendNotificationAndEmail(userID uint, title, message string) {
	// Send notification
	s.notificationService.DispatchNotification(
		userID,
		title,
		message,
		string(enums.NotificationTypeMarketplace),
	)

	// Get user email
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		s.logger.Errorw("Failed to get user for email notification",
			"error", err,
			"userID", userID,
		)
		return
	}

	// Send email with simple template (for backward compatibility)
	_, err = s.emailClient.SendEmail(&email.SendEmailParams{
		From:    "noreply@notifications.hellobutter.io",
		To:      []string{user.Email},
		Subject: title,
		Html:    fmt.Sprintf("<p>%s</p>", message),
		Text:    message,
	})
	if err != nil {
		s.logger.Errorw("Failed to send email notification",
			"error", err,
			"userID", userID,
			"email", user.Email,
		)
	}
}

// sendNotificationAndEmailWithTemplate sends notification and email with rich HTML template
func (s *marketplaceOrderManagementServiceImpl) sendNotificationAndEmailWithTemplate(userID uint, title, notificationMessage, subject, htmlBody, textBody string) {
	// Send notification (keep short for in-app)
	s.notificationService.DispatchNotification(
		userID,
		title,
		notificationMessage,
		string(enums.NotificationTypeMarketplace),
	)

	// Get user email
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		s.logger.Errorw("Failed to get user for email notification",
			"error", err,
			"userID", userID,
		)
		return
	}

	// Send rich HTML email
	_, err = s.emailClient.SendEmail(&email.SendEmailParams{
		From:    "noreply@notifications.hellobutter.io",
		To:      []string{user.Email},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	})
	if err != nil {
		s.logger.Errorw("Failed to send email notification",
			"error", err,
			"userID", userID,
			"email", user.Email,
		)
	}
}

// DeleteClientMarketplaceOrders deletes all marketplace orders and their related data for a client in a transaction.
// This method should be called within an existing transaction (tx) to ensure atomicity with other operations.
