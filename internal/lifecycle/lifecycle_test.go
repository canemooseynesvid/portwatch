package lifecycle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/portwatch/internal/lifecycle"
)

func TestOnStart_RunsHooksInOrder(t *testing.T) {
	m := lifecycle.New(time.Second)
	var order []string
	m.OnStart("a", func(_ context.Context) error { order = append(order, "a"); return nil })
	m.OnStart("b", func(_ context.Context) error { order = append(order, "b"); return nil })

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so Run exits after start
	_ = m.Run(ctx)

	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestOnStop_RunsHooksLIFO(t *testing.T) {
	m := lifecycle.New(time.Second)
	var order []string
	m.OnStop("first", func(_ context.Context) error { order = append(order, "first"); return nil })
	m.OnStop("second", func(_ context.Context) error { order = append(order, "second"); return nil })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = m.Run(ctx)

	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Fatalf("expected LIFO order, got: %v", order)
	}
}

func TestStart_ErrorAbortsRemainingHooks(t *testing.T) {
	m := lifecycle.New(time.Second)
	ran := false
	m.OnStart("fail", func(_ context.Context) error { return errors.New("boom") })
	m.OnStart("never", func(_ context.Context) error { ran = true; return nil })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := m.Run(ctx)

	if err == nil {
		t.Fatal("expected error from failing start hook")
	}
	if ran {
		t.Fatal("second hook should not have run after failure")
	}
}

func TestStop_ContinuesOnError(t *testing.T) {
	m := lifecycle.New(time.Second)
	stopped := false
	m.OnStop("err", func(_ context.Context) error { return errors.New("stop error") })
	m.OnStop("ok", func(_ context.Context) error { stopped = true; return nil })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.Run(ctx)

	if err == nil {
		t.Fatal("expected error propagated from stop hook")
	}
	if !stopped {
		t.Fatal("second stop hook should still have run")
	}
}

func TestNew_DefaultShutdownTimeout(t *testing.T) {
	m := lifecycle.New(5 * time.Second)
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}
