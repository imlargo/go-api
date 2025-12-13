package enums

type OnlyfansRevenueType string

const (
	OnlyfansRevenueTypeRecurringSubscription OnlyfansRevenueType = "recurring_subscription"
	OnlyfansRevenueTypeSubscription          OnlyfansRevenueType = "subscription"
	OnlyfansRevenueTypeTip                   OnlyfansRevenueType = "tip"
	OnlyfansRevenueTypeMessage               OnlyfansRevenueType = "message"
	OnlyfansRevenueTypePost                  OnlyfansRevenueType = "post"
)

func (o OnlyfansRevenueType) IsValid() bool {
	switch o {
	case OnlyfansRevenueTypeRecurringSubscription,
		OnlyfansRevenueTypeSubscription,
		OnlyfansRevenueTypeTip,
		OnlyfansRevenueTypeMessage,
		OnlyfansRevenueTypePost:
		return true
	default:
		return false
	}
}

func (o OnlyfansRevenueType) Label() string {
	switch o {
	case OnlyfansRevenueTypeRecurringSubscription:
		return "Recurring Subscription"
	case OnlyfansRevenueTypeSubscription:
		return "Subscription"
	case OnlyfansRevenueTypeTip:
		return "Tip"
	case OnlyfansRevenueTypeMessage:
		return "Paid Message"
	case OnlyfansRevenueTypePost:
		return "Post"
	default:
		return "Unknown"
	}
}

type OnlyfansAuthStatus string

const (
	OnlyfansAuthStatusPending      OnlyfansAuthStatus = "pending"
	OnlyfansAuthStatusConnected    OnlyfansAuthStatus = "connected"
	OnlyfansAuthStatusDisconnected OnlyfansAuthStatus = "disconnected"
)

func (o OnlyfansAuthStatus) IsValid() bool {
	switch o {
	case OnlyfansAuthStatusPending,
		OnlyfansAuthStatusConnected,
		OnlyfansAuthStatusDisconnected:
		return true
	default:
		return false
	}
}

// OnlyfansWebhookEvent represents the type of webhook event from OnlyFans
type OnlyfansWebhookEvent string

const (
	OnlyfansWebhookEventMessagesPpvUnlocked  OnlyfansWebhookEvent = "messages.ppv.unlocked"
	OnlyfansWebhookEventSubscriptionsNew     OnlyfansWebhookEvent = "subscriptions.new"
	OnlyfansWebhookEventAuthenticationFailed OnlyfansWebhookEvent = "accounts.authentication_failed"
)

func (e OnlyfansWebhookEvent) IsValid() bool {
	switch e {
	case OnlyfansWebhookEventMessagesPpvUnlocked,
		OnlyfansWebhookEventSubscriptionsNew,
		OnlyfansWebhookEventAuthenticationFailed:
		return true
	default:
		return false
	}
}
