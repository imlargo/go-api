package push

import (
	"encoding/json"

	"github.com/SherClockHolmes/webpush-go"
)

type PushNotifier interface {
	Send(subscription *Subscription, payload interface{}) error
}

type pushNotifier struct {
	vapidPrivateKey string
	vapidPublicKey  string
}

func NewPushNotifier(vapidPrivateKey string, vapidPublicKey string) PushNotifier {
	return &pushNotifier{
		vapidPrivateKey: vapidPrivateKey,
		vapidPublicKey:  vapidPublicKey,
	}
}

func (p *pushNotifier) Send(subscription *Subscription, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	sub := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256dh,
			Auth:   subscription.Auth,
		},
	}

	resp, err := webpush.SendNotification(payloadBytes, sub, &webpush.Options{
		Subscriber:      "https://app.hellobutter.io",
		VAPIDPublicKey:  p.vapidPublicKey,
		VAPIDPrivateKey: p.vapidPrivateKey,
		TTL:             30,
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
