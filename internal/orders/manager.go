
package orders

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rajnish-012/delivery-management-system/internal/database"
	"github.com/rajnish-012/delivery-management-system/internal/models"
)

// orderController tracks a running progression goroutine for an order
type orderController struct {
	stop chan struct{}
}

var (
	controllers = make(map[int]*orderController) // orderID -> controller
	mu          sync.Mutex
)

// lifecycle defines the ordered states an order goes through
var lifecycle = []string{"created", "dispatched", "in_transit", "delivered"}

// StartProgression launches a goroutine to move the order through lifecycle states.
// It is safe to call StartProgression multiple times; only one goroutine per order will run.
func StartProgression(ctx context.Context, orderID int) {
	mu.Lock()
	// Already running?
	if _, ok := controllers[orderID]; ok {
		mu.Unlock()
		return
	}
	c := &orderController{stop: make(chan struct{})}
	controllers[orderID] = c
	mu.Unlock()

	go func() {
		defer func() {
			mu.Lock()
			delete(controllers, orderID)
			mu.Unlock()
		}()

		// fetch current status
		ord, err := models.GetOrderByID(ctx, orderID)
		if err != nil {
			return
		}

		// locate index in lifecycle
		idx := -1
		for i, s := range lifecycle {
			if s == ord.Status {
				idx = i
				break
			}
		}
		if idx == -1 {
			// unknown status -> start from beginning
			idx = 0
			// ensure DB is consistent
			_ = models.UpdateOrderStatus(ctx, orderID, lifecycle[idx])
			publishUpdate(ctx, orderID, lifecycle[idx])
		}

		// progression loop
		for idx < len(lifecycle)-1 {
			// Wait time between state transitions.
			// For production/testing, expose this via config/env.
			select {
			case <-time.After(5 * time.Second):
				// Re-check order status from DB (in case it was cancelled externally)
				ordNow, err := models.GetOrderByID(ctx, orderID)
				if err != nil {
					return
				}
				if ordNow.Status == "cancelled" {
					// publish that it's cancelled and stop progressing
					publishUpdate(ctx, orderID, "cancelled")
					return
				}

				// advance to next state
				idx++
				nextStatus := lifecycle[idx]
				if err := models.UpdateOrderStatus(ctx, orderID, nextStatus); err != nil {
					// if update fails, stop progression
					return
				}
				publishUpdate(ctx, orderID, nextStatus)

			case <-c.stop:
				// stopped by cancellation/override
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// CancelProgression stops any running progression goroutine for an order.
func CancelProgression(orderID int) {
	mu.Lock()
	defer mu.Unlock()
	if c, ok := controllers[orderID]; ok {
		// closing stop channel signals goroutine to terminate
		close(c.stop)
		delete(controllers, orderID)
	}
}

// publishUpdate publishes a JSON payload to Redis channel orders:updates
func publishUpdate(ctx context.Context, orderID int, status string) {
	if database.Rdb != nil {
		payload := fmt.Sprintf(`{"order_id":%d,"status":"%s"}`, orderID, status)
		_ = database.Rdb.Publish(ctx, "orders:updates", payload).Err()
	}
}

// PublishImmediateUpdate is an exported helper to publish updates (used by handlers)
func PublishImmediateUpdate(ctx context.Context, orderID int, status string) {
	publishUpdate(ctx, orderID, status)
}
