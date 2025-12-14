package pubsub

import (
	"fmt"
	"sync"
	"time"
)

var idCounter uint64
var idMutex sync.Mutex

func generateID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter++
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), idCounter)
}
