package sse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/imlargo/go-api-template/internal/domain/models"
	"github.com/imlargo/go-api-template/internal/shared/ports"
)

type SubscriptionManager struct {
	clients    map[string]*ConnectionClient
	userIndex  map[uint]map[string]*ConnectionClient // userID -> deviceID -> Client
	mutex      sync.RWMutex
	pingTicker *time.Ticker
}

func NewNotificationSubscriptionManager() ports.SSENotificationDispatcher {
	service := &SubscriptionManager{
		clients:    make(map[string]*ConnectionClient),
		userIndex:  make(map[uint]map[string]*ConnectionClient),
		pingTicker: time.NewTicker(30 * time.Second),
	}

	// Cleanup routine for dead connections
	go service.cleanupRoutine()

	return service
}

func (sm *SubscriptionManager) Subscribe(ctx context.Context, userID uint, deviceID string) (ports.SSENotificationConnection, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if existingClient, exists := sm.clients[deviceID]; exists {
		// Cancelar la conexión existente
		existingClient.Cancel()
		sm.removeClientUnsafe(deviceID)
	}

	clientCtx, cancel := context.WithCancel(ctx)

	client := &ConnectionClient{
		ID:       deviceID,
		UserID:   userID,
		Channel:  make(chan *models.Notification, 100), // Buffer para evitar bloqueos
		Context:  clientCtx,
		Cancel:   cancel,
		LastSeen: time.Now(),
	}

	sm.clients[deviceID] = client

	if sm.userIndex[userID] == nil {
		sm.userIndex[userID] = make(map[string]*ConnectionClient)
	}

	sm.userIndex[userID][deviceID] = client

	return client, nil
}

// Unsubscribe desuscribe un dispositivo
func (sm *SubscriptionManager) Unsubscribe(userID uint, deviceID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	client, exists := sm.clients[deviceID]
	if !exists {
		return fmt.Errorf("client not found: %s", deviceID)
	}

	if client.UserID != userID {
		return fmt.Errorf("userID does not match for device: %s", deviceID)
	}

	client.Cancel()
	sm.removeClientUnsafe(deviceID)

	return nil
}

func (sm *SubscriptionManager) Send(notification *models.Notification) error {
	sm.mutex.RLock()
	userClients, exists := sm.userIndex[notification.UserID]
	if !exists {
		sm.mutex.RUnlock()
		return fmt.Errorf("no subscribed devices for user: %d", notification.UserID)
	}

	// Prevent concurrentcy issues by copying the map
	clients := make([]*ConnectionClient, 0, len(userClients))
	for _, client := range userClients {
		clients = append(clients, client)
	}
	sm.mutex.RUnlock()

	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)

		go func(c *ConnectionClient) {
			defer wg.Done()
			select {
			case c.Channel <- notification:
				// Notification sent
			case <-time.After(5 * time.Second):
				// Timeout sending notification
			case <-c.Context.Done():
				// Client disconnected during send
			}
		}(client)
	}

	wg.Wait()

	return nil
}

// removeClientUnsafe remueve un cliente (debe llamarse con mutex bloqueado)
func (sm *SubscriptionManager) removeClientUnsafe(deviceID string) {
	client, exists := sm.clients[deviceID]
	if !exists {
		return
	}

	// Cerrar canal
	close(client.Channel)

	// Remover de índices
	delete(sm.clients, deviceID)
	if userClients, exists := sm.userIndex[client.UserID]; exists {
		delete(userClients, deviceID)
		if len(userClients) == 0 {
			delete(sm.userIndex, client.UserID)
		}
	}
}

func (sm *SubscriptionManager) cleanupRoutine() {
	for range sm.pingTicker.C {
		sm.mutex.Lock()
		now := time.Now()
		var toRemove []string

		for deviceID, client := range sm.clients {
			select {
			case <-client.Context.Done():
				toRemove = append(toRemove, deviceID)
			default:
				// Verificar si la conexión está muy antigua
				if now.Sub(client.LastSeen) > 2*time.Minute {
					client.Cancel()
					toRemove = append(toRemove, deviceID)
				}
			}
		}

		for _, deviceID := range toRemove {
			sm.removeClientUnsafe(deviceID)
		}

		sm.mutex.Unlock()
	}
}

func (sm *SubscriptionManager) GetSSESubscriptions() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	userCount := len(sm.userIndex)
	deviceCount := len(sm.clients)

	return map[string]interface{}{
		"users":   userCount,
		"devices": deviceCount,
	}
}
