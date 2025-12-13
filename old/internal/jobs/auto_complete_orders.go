package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
)

type AutoCompleteOrders struct {
	store               *store.Store
	notificationService services.NotificationService
}

func NewAutoCompleteOrdersTask(
	store *store.Store,
	notificationService services.NotificationService,
) Job {
	return &AutoCompleteOrders{
		store:               store,
		notificationService: notificationService,
	}
}

// AutoCompleteOrders is a job responsible for automatically completing orders
// that have been delivered and not reviewed within the auto-completion period.
// It retrieves all orders with 'delivered' status where the auto-completion date
// has passed, updates them to 'auto_completed' status, and sends notifications
// to both buyer and seller. This ensures that sellers receive payment for work
// that has been delivered and not disputed within the review period.
func (a *AutoCompleteOrders) Execute() error {
	// Get all orders that are eligible for auto-completion
	ordersToComplete, err := a.store.MarketplaceOrders.GetOrdersForAutoCompletion()
	if err != nil {
		return fmt.Errorf("error fetching orders for auto-completion: %w", err)
	}

	log.Printf("Found %d orders to auto-complete\n", len(ordersToComplete))

	for i, order := range ordersToComplete {
		fmt.Printf("Auto-completing order %d/%d: ID#%d\n", i+1, len(ordersToComplete), order.ID)

		// Update order status
		order.Status = enums.MarketplaceOrderStatusAutoCompleted
		order.CompletedAt = time.Now()

		if err := a.store.MarketplaceOrders.Update(order); err != nil {
			log.Printf("Error auto-completing order ID#%d: %v\n", order.ID, err)
			continue
		}

		// Add timeline event
		timeline := &models.MarketplaceOrderTimeline{
			OrderID:     order.ID,
			EventType:   enums.OrderTimelineEventCompleted,
			Description: "Order auto-completed after review period",
		}
		if err := a.store.MarketplaceOrderTimelines.Create(timeline); err != nil {
			log.Printf("Error creating timeline event for order ID#%d: %v\n", order.ID, err)
		}

		// Notify buyer
		a.notificationService.DispatchNotification(
			order.BuyerID,
			"Order Auto-Completed",
			fmt.Sprintf("Order #%d has been automatically completed. Payment has been released to the seller.", order.ID),
			string(enums.NotificationTypeMarketplace),
		)

		// Notify seller
		if order.Service != nil {
			a.notificationService.DispatchNotification(
				order.Service.UserID,
				"Order Auto-Completed",
				fmt.Sprintf("Order #%d has been automatically completed. Payment has been released to you.", order.ID),
				string(enums.NotificationTypeMarketplace),
			)
		}
	}

	log.Printf("Successfully auto-completed %d orders\n", len(ordersToComplete))
	return nil
}

func (a *AutoCompleteOrders) GetName() TaskLabel {
	return TaskAutoCompleteOrders
}
