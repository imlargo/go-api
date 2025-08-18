package ports

import "github.com/SherClockHolmes/webpush-go"

type PushNotifier interface {
	SendNotification(subscription *webpush.Subscription, payload interface{}) error
}
