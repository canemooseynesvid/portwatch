package audit

import (
	"github.com/user/portwatch/internal/alerting"
)

// AlertAdapter bridges the alerting pipeline to the audit Logger.
// It satisfies the alerting.Handler interface.
type AlertAdapter struct {
	logger *Logger
}

// NewAlertAdapter wraps logger so it can be used as an alerting.Handler.
func NewAlertAdapter(logger *Logger) *AlertAdapter {
	return &AlertAdapter{logger: logger}
}

// Handle converts an alerting.Alert into an audit Event and records it.
func (a *AlertAdapter) Handle(al alerting.Alert) {
	lvl := levelFromAlert(al)
	ev := Event{
		Timestamp: al.Timestamp,
		Level:     lvl,
		Message:   al.Message,
		Port:      al.Port,
		Protocol:  al.Protocol,
		PID:       al.PID,
	}
	// Best-effort: ignore write errors so a broken log file never
	// disrupts the monitoring loop.
	_ = a.logger.Record(ev) //nolint:errcheck
}

func levelFromAlert(al alerting.Alert) Level {
	switch al.Level {
	case alerting.LevelInfo:
		return LevelInfo
	case alerting.LevelWarn:
		return LevelWarn
	default:
		return LevelAlert
	}
}
