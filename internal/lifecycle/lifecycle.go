// Package lifecycle manages graceful startup and shutdown of portwatch daemon components.
package lifecycle

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Hook is a function called during a lifecycle phase.
type Hook func(ctx context.Context) error

// Manager coordinates startup and shutdown of registered components.
type Manager struct {
	startHooks  []namedHook
	stopHooks   []namedHook
	shutdownTO  time.Duration
	mu          sync.Mutex
}

type namedHook struct {
	name string
	fn   Hook
}

// New returns a Manager with the given graceful shutdown timeout.
func New(shutdownTimeout time.Duration) *Manager {
	return &Manager{shutdownTO: shutdownTimeout}
}

// OnStart registers a hook to run during startup.
func (m *Manager) OnStart(name string, fn Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startHooks = append(m.startHooks, namedHook{name, fn})
}

// OnStop registers a hook to run during shutdown (LIFO order).
func (m *Manager) OnStop(name string, fn Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopHooks = append(m.stopHooks, namedHook{name, fn})
}

// Run starts all components, blocks until a signal is received, then shuts down.
func (m *Manager) Run(ctx context.Context) error {
	if err := m.start(ctx); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Printf("lifecycle: received signal %s, shutting down", sig)
	case <-ctx.Done():
		log.Printf("lifecycle: context cancelled, shutting down")
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), m.shutdownTO)
	defer cancel()
	return m.stop(shutCtx)
}

func (m *Manager) start(ctx context.Context) error {
	for _, h := range m.startHooks {
		log.Printf("lifecycle: starting %s", h.name)
		if err := h.fn(ctx); err != nil {
			return fmt.Errorf("lifecycle: start %s: %w", h.name, err)
		}
	}
	return nil
}

func (m *Manager) stop(ctx context.Context) error {
	hooks := m.stopHooks
	var first error
	for i := len(hooks) - 1; i >= 0; i-- {
		h := hooks[i]
		log.Printf("lifecycle: stopping %s", h.name)
		if err := h.fn(ctx); err != nil && first == nil {
			first = fmt.Errorf("lifecycle: stop %s: %w", h.name, err)
		}
	}
	return first
}
