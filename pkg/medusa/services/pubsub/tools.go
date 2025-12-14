package pubsub

import (
	"fmt"
	"time"
)

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
