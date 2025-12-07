package push

type Subscription struct {
	Endpoint string `json:"endpoint"`
	Auth     string `json:"auth"`
	P256dh   string `json:"p256dh"`
}
