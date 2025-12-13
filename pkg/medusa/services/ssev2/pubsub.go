package ssev2

// PubSub defines the interface for distributed messaging
type PubSub interface {
	Publish(topic string, event Event) error
	Subscribe(topics []string, handler func(topic string, event Event)) error
	Unsubscribe(topics []string) error
	Close() error
}
