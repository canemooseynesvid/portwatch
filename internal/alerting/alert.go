package alerting

import (
	"fmt"
	"time"
)

// AlertLevel represents the severity of an alert.
type AlertLevel int

const (
	AlertInfo AlertLevel = iota
	AlertWarn
	AlertCritical
)

func (l AlertLevel) String() string {
	switch l {
	case AlertInfo:
		return "INFO"
	case AlertWarn:
		return "WARN"
	case AlertCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Alert represents a port-related alert event.
type Alert struct {
	Level     AlertLevel
	Message   string
	Port      uint16
	Protocol  string
	Timestamp time.Time
}

func (a Alert) String() string {
	return fmt.Sprintf("[%s] %s | port=%d proto=%s ts=%s",
		a.Level, a.Message, a.Port, a.Protocol, a.Timestamp.Format(time.RFC3339))
}

// Alerter dispatches alerts to one or more handlers.
type Alerter struct {
	handlers []Handler
}

// Handler is a function that receives and processes an Alert.
type Handler func(Alert)

// NewAlerter creates an Alerter with the given handlers.
func NewAlerter(handlers ...Handler) *Alerter {
	return &Alerter{handlers: handlers}
}

// Send dispatches the alert to all registered handlers.
func (a *Alerter) Send(alert Alert) {
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}
	for _, h := range a.handlers {
		h(alert)
	}
}

// NewPortBindAlert creates an alert for an unexpected port binding.
func NewPortBindAlert(port uint16, protocol string) Alert {
	return Alert{
		Level:    AlertWarn,
		Message:  fmt.Sprintf("unexpected binding detected on port %d/%s", port, protocol),
		Port:     port,
		Protocol: protocol,
	}
}

// NewPortConflictAlert creates an alert for a port conflict.
func NewPortConflictAlert(port uint16, protocol string) Alert {
	return Alert{
		Level:    AlertCritical,
		Message:  fmt.Sprintf("port conflict detected on port %d/%s", port, protocol),
		Port:     port,
		Protocol: protocol,
	}
}
