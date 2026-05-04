package grpc

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
)

type OrderUpdate struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ItemName   string    `json:"item_name"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderUpdateBroker struct {
	listener    *pq.Listener
	done        chan struct{}
	mu          sync.RWMutex
	subscribers map[string]map[chan OrderUpdate]struct{}
}

func NewOrderUpdateBroker(dsn string) (*OrderUpdateBroker, error) {
	listener := pq.NewListener(dsn, 10*time.Second, time.Minute, func(_ pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("order update listener error: %v", err)
		}
	})
	if err := listener.Listen("order_updates"); err != nil {
		return nil, err
	}

	broker := &OrderUpdateBroker{
		listener:    listener,
		done:        make(chan struct{}),
		subscribers: make(map[string]map[chan OrderUpdate]struct{}),
	}
	go broker.run()

	return broker, nil
}

func (b *OrderUpdateBroker) Subscribe(orderID string) (<-chan OrderUpdate, func()) {
	ch := make(chan OrderUpdate, 8)

	b.mu.Lock()
	if _, ok := b.subscribers[orderID]; !ok {
		b.subscribers[orderID] = make(map[chan OrderUpdate]struct{})
	}
	b.subscribers[orderID][ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subscribers, ok := b.subscribers[orderID]
		if !ok {
			return
		}
		delete(subscribers, ch)
		if len(subscribers) == 0 {
			delete(b.subscribers, orderID)
		}
	}

	return ch, unsubscribe
}

func (b *OrderUpdateBroker) Close() {
	close(b.done)
	b.mu.Lock()
	b.subscribers = make(map[string]map[chan OrderUpdate]struct{})
	b.mu.Unlock()
	_ = b.listener.UnlistenAll()
	_ = b.listener.Close()
}

func (b *OrderUpdateBroker) run() {
	for {
		select {
		case <-b.done:
			return
		case notification := <-b.listener.Notify:
			if notification == nil {
				continue
			}

			var update OrderUpdate
			if err := json.Unmarshal([]byte(notification.Extra), &update); err != nil {
				log.Printf("failed to decode order update notification: %v", err)
				continue
			}

			b.dispatch(update)
		}
	}
}

func (b *OrderUpdateBroker) dispatch(update OrderUpdate) {
	b.mu.RLock()
	subscribers := b.subscribers[update.ID]
	channels := make([]chan OrderUpdate, 0, len(subscribers))
	for ch := range subscribers {
		channels = append(channels, ch)
	}
	b.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- update:
		default:
		}
	}
}
