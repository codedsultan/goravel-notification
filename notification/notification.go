// Package notification provides the Manager implementation for codedsultan/goravel-notification.
package notification

import (
	"fmt"
	"sync"

	contractsnotification "github.com/codedsultan/goravel-notification/contracts"

	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

// Manager is the concrete implementation of contracts.Manager.
// It is instantiated once by the ServiceProvider and bound into Goravel's
// service container under the key defined in setup/setup.go.
type Manager struct {
	mu       sync.RWMutex
	channels map[string]contractsnotification.Channel
	log      log.Log
	queue    queue.Queue // may be nil when queue is not configured
}

// NewManager constructs a Manager. The queue argument may be nil; notifications
// that implement ShouldQueue will fall back to synchronous delivery in that case.
func NewManager(logger log.Log, q queue.Queue) *Manager {
	return &Manager{
		channels: make(map[string]contractsnotification.Channel),
		log:      logger,
		queue:    q,
	}
}

// Extend satisfies contracts.Manager. It is safe to call from multiple goroutines.
func (m *Manager) Extend(ch contractsnotification.Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[ch.Name()] = ch
}

// Channel satisfies contracts.Manager.
func (m *Manager) Channel(name string) (contractsnotification.Channel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ch, ok := m.channels[name]
	if !ok {
		return nil, fmt.Errorf("notifications: channel %q is not registered", name)
	}
	return ch, nil
}

// Send satisfies contracts.Manager.
// If the notification implements ShouldQueue AND a queue is configured, it is
// dispatched asynchronously; otherwise it is delivered in the calling goroutine.
func (m *Manager) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	if sq, ok := n.(contractsnotification.ShouldQueue); ok && m.queue != nil {
		return m.dispatchQueued(notifiable, n, sq)
	}
	return m.dispatchSync(notifiable, n)
}

// SendNow satisfies contracts.Manager. Always synchronous.
func (m *Manager) SendNow(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	return m.dispatchSync(notifiable, n)
}

// dispatchSync iterates over Via() channels and calls each driver's Send.
// Errors from individual channels are logged but do not abort other channels.
func (m *Manager) dispatchSync(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	channels := n.Via(notifiable)
	if len(channels) == 0 {
		m.log.Errorf("notifications: %T.Via() returned no channels for %T — nothing sent", n, notifiable)
		return nil
	}

	var firstErr error
	for _, name := range channels {
		ch, err := m.Channel(name)
		if err != nil {
			m.log.Errorf("notifications: skipping unregistered channel %q: %v", name, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		if err := ch.Send(notifiable, n); err != nil {
			m.log.Errorf("notifications: channel %q failed for %T: %v", name, n, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// dispatchQueued pushes the notification through Goravel's queue system.
// The job itself calls dispatchSync when it is consumed by a worker.
func (m *Manager) dispatchQueued(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
	sq contractsnotification.ShouldQueue,
) error {
	job := &sendNotificationJob{
		manager:    m,
		notifiable: notifiable,
		n:          n,
	}

	pending := m.queue.Job(job)
	if conn := sq.OnConnection(); conn != "" {
		pending = pending.OnConnection(conn)
	}
	if q := sq.OnQueue(); q != "" {
		pending = pending.OnQueue(q)
	}
	return pending.Dispatch()
}
