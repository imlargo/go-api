package sse2

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SSEManager interface {
	Send(userID uint, message *Message) error
	Subscribe(ctx context.Context, userID uint, clientID string) (SSENotificationConnection, error)
	Unsubscribe(userID uint, clientID string) error
	GetSSESubscriptions() map[string]interface{}
}

type sseManager struct {
	clients    map[string]*ConnectionClient
	userIndex  map[uint]map[string]*ConnectionClient // userID -> clientID -> Client
	mutex      sync.RWMutex
	pingTicker *time.Ticker
}

func NewSSEManager() SSEManager {
	service := &sseManager{
		clients:    make(map[string]*ConnectionClient),
		userIndex:  make(map[uint]map[string]*ConnectionClient),
		pingTicker: time.NewTicker(30 * time.Second),
	}

	// Cleanup routine for dead connections
	go service.cleanupRoutine()

	return service
}

func (sm *sseManager) Subscribe(ctx context.Context, userID uint, clientID string) (SSENotificationConnection, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if existingClient, exists := sm.clients[clientID]; exists {
		// Cancelar la conexión existente
		existingClient.Cancel()
		sm.removeClientUnsafe(clientID)
	}

	clientCtx, cancel := context.WithCancel(ctx)

	client := &ConnectionClient{
		ID:       clientID,
		UserID:   userID,
		Channel:  make(chan *Message, 100), // Buffer para evitar bloqueos
		Context:  clientCtx,
		Cancel:   cancel,
		LastSeen: time.Now(),
	}

	sm.clients[clientID] = client

	if sm.userIndex[userID] == nil {
		sm.userIndex[userID] = make(map[string]*ConnectionClient)
	}

	sm.userIndex[userID][clientID] = client

	return client, nil
}

// Unsubscribe desuscribe un dispositivo
func (sm *sseManager) Unsubscribe(userID uint, clientID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	client, exists := sm.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	if client.UserID != userID {
		return fmt.Errorf("userID does not match for client: %s", clientID)
	}

	client.Cancel()
	sm.removeClientUnsafe(clientID)

	return nil
}

func (sm *sseManager) Send(userID uint, message *Message) error {
	sm.mutex.RLock()
	userClients, exists := sm.userIndex[userID]
	if !exists {
		sm.mutex.RUnlock()
		return fmt.Errorf("no subscribed clients for user: %d", userID)
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
			case c.Channel <- message:
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
func (sm *sseManager) removeClientUnsafe(clientID string) {
	client, exists := sm.clients[clientID]
	if !exists {
		return
	}

	// Cerrar canal
	close(client.Channel)

	// Remover de índices
	delete(sm.clients, clientID)
	if userClients, exists := sm.userIndex[client.UserID]; exists {
		delete(userClients, clientID)
		if len(userClients) == 0 {
			delete(sm.userIndex, client.UserID)
		}
	}
}

func (sm *sseManager) cleanupRoutine() {
	for range sm.pingTicker.C {
		sm.mutex.Lock()
		now := time.Now()
		var toRemove []string

		for clientID, client := range sm.clients {
			select {
			case <-client.Context.Done():
				toRemove = append(toRemove, clientID)
			default:
				// Verificar si la conexión está muy antigua
				if now.Sub(client.LastSeen) > 2*time.Minute {
					client.Cancel()
					toRemove = append(toRemove, clientID)
				}
			}
		}

		for _, clientID := range toRemove {
			sm.removeClientUnsafe(clientID)
		}

		sm.mutex.Unlock()
	}
}

func (sm *sseManager) GetSSESubscriptions() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	userCount := len(sm.userIndex)
	deviceCount := len(sm.clients)

	return map[string]interface{}{
		"users":   userCount,
		"devices": deviceCount,
	}
}
