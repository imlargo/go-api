package ssev2

// Subscription manages topic subscriptions for a client
type Subscription struct {
	ClientID string
	Topics   []string
}
