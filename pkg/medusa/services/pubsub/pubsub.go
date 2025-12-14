package pubsub

// PubSub combines both Publisher and Subscriber interfaces
type PubSub interface {
	Publisher
	Subscriber
}
