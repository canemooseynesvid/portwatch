package classify

import (
	"fmt"

	"github.com/example/portwatch/internal/alerting"
	"github.com/example/portwatch/internal/portscanner"
)

// TagMiddleware wraps an alerting.Handler and injects a "category" tag
// derived from the port number embedded in the alert's metadata.
type TagMiddleware struct {
	classifier *Classifier
	next       alerting.Handler
}

// NewTagMiddleware returns a TagMiddleware that tags alerts before forwarding.
func NewTagMiddleware(next alerting.Handler) *TagMiddleware {
	return &TagMiddleware{
		classifier: New(),
		next:       next,
	}
}

// Handle tags the alert with its port category and forwards it.
func (m *TagMiddleware) Handle(a alerting.Alert) error {
	if m.next == nil {
		return nil
	}
	if port, ok := a.Meta["port"]; ok {
		var p uint16
		if _, err := fmt.Sscanf(fmt.Sprintf("%v", port), "%d", &p); err == nil {
			e := portscanner.Entry{Port: p}
			cat := m.classifier.Categorize(e)
			if a.Meta == nil {
				a.Meta = make(map[string]any)
			}
			a.Meta["category"] = cat.String()
		}
	}
	return m.next.Handle(a)
}
