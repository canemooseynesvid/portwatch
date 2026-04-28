package alerting

import "fmt"

// NewPortBindAlert creates an alert for a newly detected port binding.
func NewPortBindAlert(port uint16, protocol string, level Level) Alert {
	return Alert{
		Level:   level,
		Message: fmt.Sprintf("new binding detected: %s port %d", protocol, port),
		Details: map[string]string{
			"port":     fmt.Sprintf("%d", port),
			"protocol": protocol,
			"event":    "bind",
		},
	}
}

// NewPortClosedAlert creates an alert for a port that is no longer bound.
func NewPortClosedAlert(port uint16, protocol string) Alert {
	return Alert{
		Level:   Info,
		Message: fmt.Sprintf("port released: %s port %d", protocol, port),
		Details: map[string]string{
			"port":     fmt.Sprintf("%d", port),
			"protocol": protocol,
			"event":    "close",
		},
	}
}

// NewConflictAlert creates a warning alert when two processes bind the same port.
func NewConflictAlert(port uint16, protocol string, pid1, pid2 int) Alert {
	return Alert{
		Level: Warning,
		Message: fmt.Sprintf(
			"port conflict on %s:%d between PID %d and PID %d",
			protocol, port, pid1, pid2,
		),
		Details: map[string]string{
			"port":     fmt.Sprintf("%d", port),
			"protocol": protocol,
			"pid1":     fmt.Sprintf("%d", pid1),
			"pid2":     fmt.Sprintf("%d", pid2),
			"event":    "conflict",
		},
	}
}
