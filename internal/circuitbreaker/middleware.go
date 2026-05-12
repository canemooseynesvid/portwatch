package circuitbreaker

import (
	"fmt"
	"time"

	"portwatch/internal/alerting"
)

// AlertMiddleware wraps an alerting.Handler and gates delivery through a
// per-key circuit breaker, preventing alert storms when a handler is unhealthy.
type AlertMiddleware struct {
	next      alerting.Handler
	breakers  map[string]*Breaker
	threshold int
	reset     time.Duration
	clock     func() time.Time
}

// NewAlertMiddleware returns an AlertMiddleware that opens the circuit after
// threshold consecutive failures per alert key and resets after resetTimeout.
func NewAlertMiddleware(next alerting.Handler, threshold int, resetTimeout time.Duration) *AlertMiddleware {
	return &AlertMiddleware{
		next:      next,
		breakers:  make(map[string]*Breaker),
		threshold: threshold,
		reset:     resetTimeout,
		clock:     time.Now,
	}
}

// Handle delivers the alert through the circuit breaker for its key.
// If the circuit is open the alert is silently dropped.
func (m *AlertMiddleware) Handle(a alerting.Alert) error {
	key := alertKey(a)
	b := m.breaker(key)

	if !b.Allow() {
		return nil
	}

	err := m.next.Handle(a)
	if err != nil {
		b.RecordFailure()
		return err
	}
	b.RecordSuccess()
	return nil
}

// BreakerState returns the current State for the given alert key.
// Useful for health checks and diagnostics.
func (m *AlertMiddleware) BreakerState(key string) State {
	if b, ok := m.breakers[key]; ok {
		return b.State()
	}
	return StateClosed
}

func (m *AlertMiddleware) breaker(key string) *Breaker {
	if b, ok := m.breakers[key]; ok {
		return b
	}
	b := NewWithClock(m.threshold, m.reset, m.clock)
	m.breakers[key] = b
	return b
}

func alertKey(a alerting.Alert) string {
	return fmt.Sprintf("%s:%s", a.Level, a.Title)
}
