package enums

type MarketplaceOrderStatus string

const (
	MarketplaceOrderStatusPendingPayment MarketplaceOrderStatus = "pending_payment"
	MarketplaceOrderStatusPaid           MarketplaceOrderStatus = "paid"
	MarketplaceOrderStatusPaymentExpired MarketplaceOrderStatus = "payment_expired"
	MarketplaceOrderStatusWaitingInfo    MarketplaceOrderStatus = "waiting_info"
	MarketplaceOrderStatusInProgress     MarketplaceOrderStatus = "in_progress"
	MarketplaceOrderStatusDelivered      MarketplaceOrderStatus = "delivered"
	MarketplaceOrderStatusInRevision     MarketplaceOrderStatus = "in_revision"
	MarketplaceOrderStatusCompleted      MarketplaceOrderStatus = "completed"
	MarketplaceOrderStatusAutoCompleted  MarketplaceOrderStatus = "auto_completed"
	MarketplaceOrderStatusDisputed       MarketplaceOrderStatus = "disputed"
	MarketplaceOrderStatusRefunded       MarketplaceOrderStatus = "refunded"
	MarketplaceOrderStatusCancelled      MarketplaceOrderStatus = "cancelled"
)

type DeliverableStatus string

const (
	DeliverableStatusPending  DeliverableStatus = "pending"
	DeliverableStatusAccepted DeliverableStatus = "accepted"
	DeliverableStatusRejected DeliverableStatus = "rejected"
)

type RevisionRequestStatus string

const (
	RevisionRequestStatusPending   RevisionRequestStatus = "pending"
	RevisionRequestStatusAccepted  RevisionRequestStatus = "accepted"
	RevisionRequestStatusCompleted RevisionRequestStatus = "completed"
)

type DisputeStatus string

const (
	DisputeStatusOpen        DisputeStatus = "open"
	DisputeStatusUnderReview DisputeStatus = "under_review"
	DisputeStatusResolved    DisputeStatus = "resolved"
	DisputeStatusClosed      DisputeStatus = "closed"
)

type DisputeReasonCategory string

const (
	DisputeReasonQualityIssues  DisputeReasonCategory = "quality_issues"
	DisputeReasonIncompleteWork DisputeReasonCategory = "incomplete_work"
	DisputeReasonLateDelivery   DisputeReasonCategory = "late_delivery"
	DisputeReasonNotAsDescribed DisputeReasonCategory = "not_as_described"
	DisputeReasonOther          DisputeReasonCategory = "other"
)

func (d DisputeReasonCategory) IsValid() bool {
	switch d {
	case DisputeReasonQualityIssues,
		DisputeReasonIncompleteWork,
		DisputeReasonLateDelivery,
		DisputeReasonNotAsDescribed,
		DisputeReasonOther:
		return true
	}
	return false
}

type DisputeResolution string

const (
	DisputeResolutionBuyerFavor  DisputeResolution = "buyer_favor"
	DisputeResolutionSellerFavor DisputeResolution = "seller_favor"
	DisputeResolutionSplit       DisputeResolution = "split"
)

type OrderTimelineEventType string

const (
	OrderTimelineEventCreated           OrderTimelineEventType = "order_created"
	OrderTimelineEventPaymentConfirmed  OrderTimelineEventType = "payment_confirmed"
	OrderTimelineEventStatusChanged     OrderTimelineEventType = "status_changed"
	OrderTimelineEventDeliverySubmitted OrderTimelineEventType = "delivery_submitted"
	OrderTimelineEventDeliveryReviewed  OrderTimelineEventType = "delivery_reviewed"
	OrderTimelineEventRevisionRequested OrderTimelineEventType = "revision_requested"
	OrderTimelineEventDisputeOpened     OrderTimelineEventType = "dispute_opened"
	OrderTimelineEventDisputeResolved   OrderTimelineEventType = "dispute_resolved"
	OrderTimelineEventDeadlineExtended  OrderTimelineEventType = "deadline_extended"
	OrderTimelineEventCompleted         OrderTimelineEventType = "order_completed"
	OrderTimelineEventCancelled         OrderTimelineEventType = "order_cancelled"
	OrderTimelineEventMessageSent       OrderTimelineEventType = "message_sent"
)
