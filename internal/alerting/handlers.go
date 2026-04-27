package alerting

import (
	"fmt"
	"io"
	"os"
)

// StdoutHandler returns a Handler that writes alerts to stdout.
func StdoutHandler() Handler {
	return WriterHandler(os.Stdout)
}

// WriterHandler returns a Handler that writes alerts to the given writer.
func WriterHandler(w io.Writer) Handler {
	return func(a Alert) {
		fmt.Fprintln(w, a.String())
	}
}

// FilterHandler wraps a Handler and only forwards alerts at or above the given level.
func FilterHandler(minLevel AlertLevel, next Handler) Handler {
	return func(a Alert) {
		if a.Level >= minLevel {
			next(a)
		}
	}
}

// CollectorHandler returns a Handler and a pointer to the slice it collects into.
// Useful for testing.
func CollectorHandler() (Handler, *[]Alert) {
	var collected []Alert
	h := func(a Alert) {
		collected = append(collected, a)
	}
	return h, &collected
}
