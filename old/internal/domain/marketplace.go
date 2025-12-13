package domain

// Marketplace order management constants
const (
	// AutoCompletionDays defines the number of days before an order is automatically completed after delivery
	AutoCompletionDays = 3

	// MaxRevisionsDefault defines the default number of revisions allowed per order when not specified in package
	MaxRevisionsDefault = 2

	// MaxDeadlineExtensionDays defines the maximum number of days that can be requested for a deadline extension
	MaxDeadlineExtensionDays = 30

	// DisputeResolutionDeadlineDays defines the target number of days for resolving a dispute
	DisputeResolutionDeadlineDays = 7

	// DeliveryDescriptionMinLength defines the minimum character length for delivery descriptions
	DeliveryDescriptionMinLength = 20

	// DisputeDescriptionMinLength defines the minimum character length for dispute descriptions
	DisputeDescriptionMinLength = 100

	// RevisionReasonMinLength defines the minimum character length for revision request reasons
	RevisionReasonMinLength = 50

	// CancellationFullRefundHours defines the time window in hours for full refund eligibility after payment
	CancellationFullRefundHours = 24
)
