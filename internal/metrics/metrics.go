package metrics

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Counter is a monotonically increasing counter.
type Counter struct {
	value uint64
}

func (c *Counter) Inc() { atomic.AddUint64(&c.value, 1) }
func (c *Counter) Add(n uint64) { atomic.AddUint64(&c.value, n) }
func (c *Counter) Value() uint64 { return atomic.LoadUint64(&c.value) }

// Registry holds named counters for portwatch runtime metrics.
type Registry struct {
	mu       sync.RWMutex
	counters map[string]*Counter
	start    time.Time
}

// New creates an empty metrics Registry.
func New() *Registry {
	return &Registry{
		counters: make(map[string]*Counter),
		start:    time.Now(),
	}
}

// Counter returns (or creates) the named counter.
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{}
	r.counters[name] = c
	return c
}

// Snapshot returns a point-in-time copy of all counter values.
func (r *Registry) Snapshot() map[string]uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]uint64, len(r.counters))
	for k, c := range r.counters {
		out[k] = c.Value()
	}
	return out
}

// Uptime returns the duration since the registry was created.
func (r *Registry) Uptime() time.Duration {
	return time.Since(r.start)
}

// Print writes a human-readable metrics summary to w.
func (r *Registry) Print(w io.Writer) {
	fmt.Fprintf(w, "uptime: %s\n", r.Uptime().Round(time.Second))
	snap := r.Snapshot()
	if len(snap) == 0 {
		fmt.Fprintln(w, "(no counters recorded)")
		return
	}
	for name, val := range snap {
		fmt.Fprintf(w, "  %-30s %d\n", name, val)
	}
}

// PrintToStdout is a convenience wrapper around Print.
func (r *Registry) PrintToStdout() {
	r.Print(os.Stdout)
}
