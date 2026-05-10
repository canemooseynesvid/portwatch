package fingerprint_test

import (
	"testing"

	"github.com/user/portwatch/internal/fingerprint"
)

func baseInfo() fingerprint.Info {
	return fingerprint.Info{
		Protocol: "tcp",
		Address:  "0.0.0.0",
		Port:     8080,
		PID:      1234,
		Comm:     "myapp",
	}
}

func TestCompute_Deterministic(t *testing.T) {
	info := baseInfo()
	a := fingerprint.Compute(info)
	b := fingerprint.Compute(info)
	if a != b {
		t.Fatalf("expected identical fingerprints, got %q and %q", a, b)
	}
}

func TestCompute_DiffersByPID(t *testing.T) {
	info := baseInfo()
	a := fingerprint.Compute(info)
	info.PID = 9999
	b := fingerprint.Compute(info)
	if a == b {
		t.Fatal("expected different fingerprints for different PIDs")
	}
}

func TestCompute_DiffersByComm(t *testing.T) {
	info := baseInfo()
	a := fingerprint.Compute(info)
	info.Comm = "otherapp"
	b := fingerprint.Compute(info)
	if a == b {
		t.Fatal("expected different fingerprints for different comm values")
	}
}

func TestChanged(t *testing.T) {
	if fingerprint.Changed("abc", "abc") {
		t.Fatal("identical fingerprints should not be changed")
	}
	if !fingerprint.Changed("abc", "xyz") {
		t.Fatal("different fingerprints should be changed")
	}
}

func TestStore_FirstObservation_NotChanged(t *testing.T) {
	s := fingerprint.NewStore()
	_, _, changed := s.Track(baseInfo())
	if changed {
		t.Fatal("first observation should not be reported as changed")
	}
	if s.Len() != 1 {
		t.Fatalf("expected Len 1, got %d", s.Len())
	}
}

func TestStore_SameProcess_NotChanged(t *testing.T) {
	s := fingerprint.NewStore()
	info := baseInfo()
	s.Track(info)
	_, _, changed := s.Track(info)
	if changed {
		t.Fatal("same info should not be reported as changed")
	}
}

func TestStore_ProcessChanged_ReportsChange(t *testing.T) {
	s := fingerprint.NewStore()
	info := baseInfo()
	s.Track(info)
	info.PID = 5555
	info.Comm = "newbinary"
	prev, cur, changed := s.Track(info)
	if !changed {
		t.Fatal("expected changed=true when process replaced")
	}
	if prev == cur {
		t.Fatal("prev and cur fingerprints should differ")
	}
}

func TestStore_Delete_RemovesEntry(t *testing.T) {
	s := fingerprint.NewStore()
	info := baseInfo()
	s.Track(info)
	s.Delete(info)
	if s.Len() != 0 {
		t.Fatalf("expected Len 0 after delete, got %d", s.Len())
	}
	// After deletion the next Track should be treated as a fresh observation.
	_, _, changed := s.Track(info)
	if changed {
		t.Fatal("re-observation after delete should not be reported as changed")
	}
}
