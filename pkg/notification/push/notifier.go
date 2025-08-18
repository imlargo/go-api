package push

import (
	"encoding/json"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/imlargo/go-api-template/internal/shared/ports"
)

type PushNotifier struct {
	vapidPrivateKey string
	vapidPublicKey  string
}

func NewPushNotifier(vapidPrivateKey string, vapidPublicKey string) ports.PushNotifier {
	return &PushNotifier{
		vapidPrivateKey: vapidPrivateKey,
		vapidPublicKey:  vapidPublicKey,
	}
}

func (p *PushNotifier) SendNotification(subscription *webpush.Subscription, payload interface{}) error {
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
