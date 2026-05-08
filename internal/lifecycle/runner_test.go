package lifecycle_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/portwatch/internal/lifecycle"
)

type fakeWorker struct {
	name    string
	blockUntilCtx bool
	errToReturn    error
	started        atomic.Bool
}

func (f *fakeWorker) Name() string { return f.name }
func (f *fakeWorker) Run(ctx context.Context) error {
	f.started.Store(true)
	if f.blockUntilCtx {
		<-ctx.Done()
	}
	return f.errToReturn
}

func TestRunner_AllWorkersStart(t *testing.T) {
	r := lifecycle.NewRunner()
	w1 := &fakeWorker{name: "w1"}
	w2 := &fakeWorker{name: "w2"}
	r.Add(w1)
	r.Add(w2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = r.Run(ctx)

	if !w1.started.Load() || !w2.started.Load() {
		t.Fatal("expected both workers to start")
	}
}

func TestRunner_ErrorCancelsOthers(t *testing.T) {
	r := lifecycle.NewRunner()
	blocker := &fakeWorker{name: "blocker", blockUntilCtx: true}
	failing := &fakeWorker{name: "failing", errToReturn: errors.New("fail")}
	r.Add(blocker)
	r.Add(failing)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := r.Run(ctx)
	if err == nil {
		t.Fatal("expected error from failing worker")
	}
}

func TestRunner_NoWorkers(t *testing.T) {
	r := lifecycle.NewRunner()
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error with no workers, got %v", err)
	}
}

func TestRunner_ContextCancelExitsCleanly(t *testing.T) {
	r := lifecycle.NewRunner()
	r.Add(&fakeWorker{name: "w", blockUntilCtx: true})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error on clean cancel: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runner did not exit after context cancel")
	}
}
