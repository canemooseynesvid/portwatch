package baseline_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/baseline"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baseline.json")
}

func TestNew_EmptyWhenMissing(t *testing.T) {
	b, err := baseline.New(tempPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(b.All()); got != 0 {
		t.Fatalf("expected 0 entries, got %d", got)
	}
}

func TestAdd_ContainsAndPersists(t *testing.T) {
	path := tempPath(t)
	b, _ := baseline.New(path)

	e := baseline.Entry{Protocol: "tcp", Address: "0.0.0.0", Port: 8080}
	if err := b.Add(e); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !b.Contains("tcp", "0.0.0.0", 8080) {
		t.Fatal("expected Contains to return true")
	}

	// Reload from disk and verify persistence.
	b2, err := baseline.New(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !b2.Contains("tcp", "0.0.0.0", 8080) {
		t.Fatal("entry not persisted across reload")
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	path := tempPath(t)
	b, _ := baseline.New(path)
	b.Add(baseline.Entry{Protocol: "udp", Address: "127.0.0.1", Port: 53})

	if err := b.Remove("udp", "127.0.0.1", 53); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if b.Contains("udp", "127.0.0.1", 53) {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	b, _ := baseline.New(tempPath(t))
	b.Add(baseline.Entry{Protocol: "tcp", Address: "0.0.0.0", Port: 443})
	b.Add(baseline.Entry{Protocol: "tcp", Address: "0.0.0.0", Port: 80})

	all := b.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the slice must not affect the baseline.
	all[0].Port = 9999
	if b.Contains("tcp", "0.0.0.0", 9999) {
		t.Fatal("mutating All() slice affected baseline")
	}
}

func TestNew_InvalidJSON(t *testing.T) {
	path := tempPath(t)
	os.WriteFile(path, []byte("not-json"), 0o600)
	_, err := baseline.New(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
