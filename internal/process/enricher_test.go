package process

import (
	"fmt"
	"testing"
)

func makeEnricher(t *testing.T) (*Enricher, *Resolver) {
	t.Helper()
	procFS := buildFakeProcFS(t, 1234, 9999)
	r := New(procFS)
	return NewEnricher(r), r
}

func TestEnrich_Found(t *testing.T) {
	procFS := buildFakeProcFS(t, 42, 7777)
	r := New(procFS)
	e := NewEnricher(r)

	ei := e.Enrich(7777)
	if !ei.Found {
		t.Fatal("expected process to be found")
	}
	if ei.Inode != 7777 {
		t.Errorf("inode mismatch: got %d, want 7777", ei.Inode)
	}
	if ei.Info.PID != 42 {
		t.Errorf("PID mismatch: got %d, want 42", ei.Info.PID)
	}
}

func TestEnrich_NotFound(t *testing.T) {
	procFS := buildFakeProcFS(t, 42, 7777)
	r := New(procFS)
	e := NewEnricher(r)

	ei := e.Enrich(9999)
	if ei.Found {
		t.Fatal("expected process not to be found")
	}
	if ei.Inode != 9999 {
		t.Errorf("inode mismatch: got %d, want 9999", ei.Inode)
	}
}

func TestEnrichMany(t *testing.T) {
	procFS := buildFakeProcFS(t, 10, 1001)
	r := New(procFS)
	e := NewEnricher(r)

	results := e.EnrichMany([]uint64{1001, 5555})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Found {
		t.Error("first entry should be found")
	}
	if results[1].Found {
		t.Error("second entry should not be found")
	}
}

func TestFormatSummary_Found(t *testing.T) {
	ei := EnrichedInfo{
		Inode: 42,
		Info:  Info{PID: 7, Comm: "nginx"},
		Found: true,
	}
	got := FormatSummary(ei)
	want := fmt.Sprintf("inode=42 pid=7 cmd=nginx")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatSummary_NotFound(t *testing.T) {
	ei := EnrichedInfo{Inode: 99, Found: false}
	got := FormatSummary(ei)
	want := "inode=99 (process not found)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
