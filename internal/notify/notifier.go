package notify

import (
	"fmt"
	"os"
	"time"
)

// Channel represents a notification delivery channel.
type Channel interface {
	Send(subject, body string) error
	Name() string
}

// Notifier dispatches notifications through one or more channels.
type Notifier struct {
	channels  []Channel
	throttle  map[string]time.Time
	cooldown  time.Duration
}

// New creates a Notifier with the given cooldown between repeated notifications.
func New(cooldown time.Duration, channels ...Channel) *Notifier {
	return &Notifier{
		channels: channels,
		throttle:  make(map[string]time.Time),
		cooldown:  cooldown,
	}
}

// Notify sends a notification through all registered channels, respecting
// the cooldown period per deduplication key.
func (n *Notifier) Notify(dedupKey, subject, body string) error {
	if last, ok := n.throttle[dedupKey]; ok {
		if time.Since(last) < n.cooldown {
			return nil
		}
	}
	n.throttle[dedupKey] = time.Now()

	var firstErr error
	for _, ch := range n.channels {
		if err := ch.Send(subject, body); err != nil {
			fmt.Fprintf(os.Stderr, "notify: channel %s error: %v\n", ch.Name(), err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Reset clears the throttle state for a given key.
func (n *Notifier) Reset(dedupKey string) {
	delete(n.throttle, dedupKey)
}
