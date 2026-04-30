package reporter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"portwatch/internal/reporter"
	"portwatch/internal/snapshot"
)

func makeEntry(proto, addr string, port, pid int) snapshot.Entry {
	return snapshot.Entry{
		Protocol:  proto,
		LocalAddr: addr,
		Port:      port,
		PID:       pid,
		State:     "LISTEN",
	}
}

func TestPrintSnapshot_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf)
	s := snapshot.New([]snapshot.Entry{
		makeEntry("tcp", "0.0.0.0", 8080, 1234),
	})
	r.PrintSnapshot(s)
	out := buf.String()
	if !strings.Contains(out, "PROTOCOL") {
		t.Error("expected PROTOCOL header in output")
	}
	if !strings.Contains(out, "8080") {
		t.Error("expected port 8080 in output")
	}
}

func TestPrintSnapshot_EmptySnapshot(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf)
	s := snapshot.New([]snapshot.Entry{})
	r.PrintSnapshot(s)
	out := buf.String()
	if !strings.Contains(out, "PROTOCOL") {
		t.Error("expected header even for empty snapshot")
	}
}

func TestPrintHistory_NoEvents(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf)
	h := snapshot.NewHistory(10)
	r.PrintHistory(h)
	out := buf.String()
	if !strings.Contains(out, "No recent events") {
		t.Errorf("expected no-events message, got: %s", out)
	}
}

func TestPrintHistory_WithEvents(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf)
	h := snapshot.NewHistory(10)
	e := makeEntry("udp", "127.0.0.1", 5353, 42)
	h.Record(snapshot.EventAdded, e, time.Now())
	r.PrintHistory(h)
	out := buf.String()
	if !strings.Contains(out, "5353") {
		t.Errorf("expected port 5353 in history output, got: %s", out)
	}
	if !strings.Contains(out, "added") {
		t.Errorf("expected event kind 'added' in output, got: %s", out)
	}
}
