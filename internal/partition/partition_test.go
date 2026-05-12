package partition_test

import (
	"errors"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/partition"
)

func makeAlert(tag string, level alerting.AlertLevel) alerting.Alert {
	return alerting.Alert{
		Level:     level,
		Message:   "test alert",
		Tag:       tag,
		Timestamp: time.Now(),
	}
}

func TestPartitioner_RoutesToCorrectBucket(t *testing.T) {
	p := partition.New(func(a alerting.Alert) string { return a.Tag }, nil)

	var got string
	p.Register("tcp", func(bucket string, _ alerting.Alert) error {
		got = bucket
		return nil
	})

	if err := p.Send(makeAlert("tcp", alerting.LevelWarning)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tcp" {
		t.Fatalf("expected bucket 'tcp', got %q", got)
	}
}

func TestPartitioner_FallbackCalledForUnknownBucket(t *testing.T) {
	var fallbackBucket string
	p := partition.New(
		func(a alerting.Alert) string { return a.Tag },
		func(bucket string, _ alerting.Alert) error {
			fallbackBucket = bucket
			return nil
		},
	)

	if err := p.Send(makeAlert("udp", alerting.LevelInfo)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fallbackBucket != "udp" {
		t.Fatalf("expected fallback bucket 'udp', got %q", fallbackBucket)
	}
}

func TestPartitioner_EmptyBucketDropsAlert(t *testing.T) {
	called := false
	p := partition.New(
		func(_ alerting.Alert) string { return "" },
		func(_ string, _ alerting.Alert) error {
			called = true
			return nil
		},
	)

	if err := p.Send(makeAlert("tcp", alerting.LevelInfo)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("fallback should not be called for empty bucket")
	}
}

func TestPartitioner_HandlerErrorPropagates(t *testing.T) {
	sentinel := errors.New("handler failure")
	p := partition.New(func(a alerting.Alert) string { return a.Tag }, nil)
	p.Register("tcp", func(_ string, _ alerting.Alert) error { return sentinel })

	err := p.Send(makeAlert("tcp", alerting.LevelCritical))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestPartitioner_BucketsReturnsRegisteredKeys(t *testing.T) {
	p := partition.New(func(a alerting.Alert) string { return a.Tag }, nil)
	p.Register("tcp", func(_ string, _ alerting.Alert) error { return nil })
	p.Register("udp", func(_ string, _ alerting.Alert) error { return nil })

	buckets := p.Buckets()
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
}
