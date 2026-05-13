// Package classify provides port-range classification for portwatch.
//
// Ports are divided into three standard IANA categories:
//
//   - System (well-known): 1–1023
//   - Registered:          1024–49151
//   - Ephemeral (dynamic): 49152–65535
//
// The Classifier type assigns a Category to any portscanner.Entry, and
// AlertLevel maps categories to suggested alerting.Level values so that
// unexpected bindings on system ports surface as warnings while registered
// or ephemeral ports emit informational alerts.
//
// TagMiddleware wraps any alerting.Handler and injects a "category" key into
// the alert's Meta map before forwarding, allowing downstream handlers and
// exporters to filter or group alerts by port class without re-computing the
// range check.
package classify
