package push

import (
	"encoding/json"

	"github.com/SherClockHolmes/webpush-go"
)

type PushNotifier interface {
	SendNotification(subscription *webpush.Subscription, payload interface{}) error
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

func (p *pushNotifier) SendNotification(subscription *webpush.Subscription, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification(payloadBytes, subscription, &webpush.Options{
		Subscriber:      "",
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
