// Package classify assigns severity categories to port binding events
// based on port ranges, protocol type, and process ownership.
package classify

import (
	"github.com/example/portwatch/internal/alerting"
	"github.com/example/portwatch/internal/portscanner"
)

// Category represents a classification tier for a port binding.
type Category int

const (
	CategoryUnknown    Category = iota
	CategorySystem              // 1–1023
	CategoryRegistered          // 1024–49151
	CategoryEphemeral           // 49152–65535
)

// String returns a human-readable label for the category.
func (c Category) String() string {
	switch c {
	case CategorySystem:
		return "system"
	case CategoryRegistered:
		return "registered"
	case CategoryEphemeral:
		return "ephemeral"
	default:
		return "unknown"
	}
}

// Classifier assigns a Category to a port entry.
type Classifier struct{}

// New returns a new Classifier.
func New() *Classifier {
	return &Classifier{}
}

// Categorize returns the Category for the given entry.
func (c *Classifier) Categorize(e portscanner.Entry) Category {
	switch {
	case e.Port >= 1 && e.Port <= 1023:
		return CategorySystem
	case e.Port >= 1024 && e.Port <= 49151:
		return CategoryRegistered
	case e.Port >= 49152 && e.Port <= 65535:
		return CategoryEphemeral
	default:
		return CategoryUnknown
	}
}

// AlertLevel returns the suggested alerting.Level for a given Category.
func AlertLevel(cat Category) alerting.Level {
	switch cat {
	case CategorySystem:
		return alerting.LevelWarning
	case CategoryRegistered:
		return alerting.LevelInfo
	case CategoryEphemeral:
		return alerting.LevelInfo
	default:
		return alerting.LevelWarning
	}
}
