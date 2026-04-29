package snapshot_test

import (
	"testing"
	"time"

	"portwatch/internal/snapshot"
)

func entry(proto, addr string, port uint16) snapshot.Entry {
	return snapshot.Entry{
		Protocol: proto,
		Address:  addr,
		Port:     port,
		PID:      1,
		SeenAt:   time.Now(),
	}
}

func TestDiff_AddedEntries(t *testing.T) {
	s := snapshot.New()
	next := []snapshot.Entry{entry("tcp", "0.0.0.0", 8080)}
	added, removed := s.Diff(next)
	if len(added) != 1 {
		t.Fatalf("expected 1 added, got %d", len(added))
	}
	if len(removed) != 0 {
		t.Fatalf("expected 0 removed, got %d", len(removed))
	}
	if added[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", added[0].Port)
	}
}

func TestDiff_RemovedEntries(t *testing.T) {
	s := snapshot.New()
	s.Set([]snapshot.Entry{entry("tcp", "0.0.0.0", 9090)})
	added, removed := s.Diff(nil)
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(removed))
	}
	if len(added) != 0 {
		t.Fatalf("expected 0 added, got %d", len(added))
	}
}

func TestDiff_NoChange(t *testing.T) {
	s := snapshot.New()
	e := entry("udp", "127.0.0.1", 53)
	s.Set([]snapshot.Entry{e})
	added, removed := s.Diff([]snapshot.Entry{e})
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("expected no diff, got added=%d removed=%d", len(added), len(removed))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := snapshot.New()
	s.Set([]snapshot.Entry{entry("tcp", "0.0.0.0", 443), entry("tcp", "0.0.0.0", 80)})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestKeyOf(t *testing.T) {
	e := entry("tcp", "0.0.0.0", 22)
	k := snapshot.KeyOf(e)
	if k.Protocol != "tcp" || k.Address != "0.0.0.0" || k.Port != 22 {
		t.Errorf("unexpected key: %+v", k)
	}
}
