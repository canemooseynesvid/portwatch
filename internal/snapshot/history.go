package snapshot

import (
	"sync"
	"time"
)

// Event records a change detected during a diff cycle.
type Event struct {
	Kind    EventKind
	Entry   Entry
	OccurredAt time.Time
}

// EventKind indicates whether a port was bound or released.
type EventKind int

const (
	EventBound   EventKind = iota // port became active
	EventReleased                 // port was closed
)

func (k EventKind) String() string {
	switch k {
	case EventBound:
		return "bound"
	case EventReleased:
		return "released"
	default:
		return "unknown"
	}
}

// History retains a bounded ring of recent Events.
type History struct {
	mu     sync.RWMutex
	events []Event
	max    int
}

// NewHistory returns a History that retains at most maxEvents entries.
func NewHistory(maxEvents int) *History {
	if maxEvents <= 0 {
		maxEvents = 256
	}
	return &History{max: maxEvents, events: make([]Event, 0, maxEvents)}
}

// Record appends an event, evicting the oldest if at capacity.
func (h *History) Record(e Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.events) >= h.max {
		h.events = h.events[1:]
	}
	h.events = append(h.events, e)
}

// Recent returns a copy of all stored events, oldest first.
func (h *History) Recent() []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Event, len(h.events))
	copy(out, h.events)
	return out
}

// Len returns the number of stored events.
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.events)
}
