// Package audit records significant portwatch events to a persistent audit log.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents the severity of an audit event.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelAlert Level = "ALERT"
)

// Event is a single audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Level     Level     `json:"level"`
	Message   string    `json:"message"`
	Port      uint16    `json:"port,omitempty"`
	Protocol  string    `json:"protocol,omitempty"`
	PID       int       `json:"pid,omitempty"`
}

// Logger writes audit events as newline-delimited JSON.
type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

// New creates a Logger that appends to the file at path.
// If path is empty, events are written to os.Stderr.
func New(path string) (*Logger, error) {
	if path == "" {
		return &Logger{out: os.Stderr}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	return &Logger{out: f}, nil
}

// NewWithWriter creates a Logger that writes to w (useful in tests).
func NewWithWriter(w io.Writer) *Logger {
	return &Logger{out: w}
}

// Record encodes ev as JSON and writes it to the underlying writer.
func (l *Logger) Record(ev Event) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	enc := json.NewEncoder(l.out)
	if err := enc.Encode(ev); err != nil {
		return fmt.Errorf("audit: encode event: %w", err)
	}
	return nil
}

// Info is a convenience wrapper for Level INFO.
func (l *Logger) Info(msg string) error {
	return l.Record(Event{Level: LevelInfo, Message: msg})
}

// Warn is a convenience wrapper for Level WARN.
func (l *Logger) Warn(msg string) error {
	return l.Record(Event{Level: LevelWarn, Message: msg})
}

// Alert is a convenience wrapper for Level ALERT.
func (l *Logger) Alert(msg string) error {
	return l.Record(Event{Level: LevelAlert, Message: msg})
}
