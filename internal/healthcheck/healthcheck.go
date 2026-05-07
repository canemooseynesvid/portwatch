package healthcheck

import (
	"fmt"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusOK      Status = "ok"
	StatusDegraded Status = "degraded"
	StatusFailed  Status = "failed"
)

// Result holds the result of a single health check.
type Result struct {
	Name      string
	Status    Status
	Message   string
	CheckedAt time.Time
}

func (r Result) String() string {
	return fmt.Sprintf("[%s] %s: %s", r.Status, r.Name, r.Message)
}

// Checker is a named health check function.
type Checker interface {
	Name() string
	Check() Result
}

// Registry holds and runs registered health checks.
type Registry struct {
	mu       sync.RWMutex
	checkers []Checker
	clock    func() time.Time
}

// New returns a new Registry.
func New() *Registry {
	return &Registry{clock: time.Now}
}

// Register adds a Checker to the registry.
func (r *Registry) Register(c Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, c)
}

// RunAll executes all registered checkers and returns their results.
func (r *Registry) RunAll() []Result {
	r.mu.RLock()
	checkers := make([]Checker, len(r.checkers))
	copy(checkers, r.checkers)
	r.mu.RUnlock()

	results := make([]Result, 0, len(checkers))
	for _, c := range checkers {
		res := c.Check()
		if res.CheckedAt.IsZero() {
			res.CheckedAt = r.clock()
		}
		results = append(results, res)
	}
	return results
}

// Overall returns the worst status across all results.
func Overall(results []Result) Status {
	worst := StatusOK
	for _, r := range results {
		switch r.Status {
		case StatusFailed:
			return StatusFailed
		case StatusDegraded:
			worst = StatusDegraded
		}
	}
	return worst
}
