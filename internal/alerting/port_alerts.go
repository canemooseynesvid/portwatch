package alerting

import "fmt"

// portDetails returns a base Details map for port-related alerts.
func portDetails(port uint16, protocol, event string) map[string]string {
	return map[string]string{
		"port":     fmt.Sprintf("%d", port),
		"protocol": protocol,
		"event":    event,
	}
}

// NewPortBindAlert creates an alert for a newly detected port binding.
func NewPortBindAlert(port uint16, protocol string, level Level) Alert {
	return Alert{
		Level:   level,
		Message: fmt.Sprintf("new binding detected: %s port %d", protocol, port),
		Details: portDetails(port, protocol, "bind"),
	}
}

// NewPortClosedAlert creates an alert for a port that is no longer bound.
func NewPortClosedAlert(port uint16, protocol string) Alert {
	return Alert{
		Level:   Info,
		Message: fmt.Sprintf("port released: %s port %d", protocol, port),
		Details: portDetails(port, protocol, "close"),
	}
}

// NewConflictAlert creates a warning alert when two processes bind the same port.
func NewConflictAlert(port uint16, protocol string, pid1, pid2 int) Alert {
	details := portDetails(port, protocol, "conflict")
	details["pid1"] = fmt.Sprintf("%d", pid1)
	details["pid2"] = fmt.Sprintf("%d", pid2)
	return Alert{
		Level: Warning,
		Message: fmt.Sprintf(
			"port conflict on %s:%d between PID %d and PID %d",
			protocol, port, pid1, pid2,
		),
		Details: details,
	}
}

// NewPrivilegedPortAlert creates a warning alert when a process binds a
// privileged port (ports 1–1023), which typically requires elevated permissions.
func NewPrivilegedPortAlert(port uint16, protocol string) Alert {
	return Alert{
		Level:   Warning,
		Message: fmt.Sprintf("privileged port bound: %s port %d", protocol, port),
		Details: portDetails(port, protocol, "privileged_bind"),
	}
}
