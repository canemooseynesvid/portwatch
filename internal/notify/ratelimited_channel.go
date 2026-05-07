package notify

import (
	"fmt"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/ratelimit"
)

// RateLimitedChannel wraps another Channel and suppresses duplicate alerts
// that exceed a configurable rate limit, emitting a summary message instead.
type RateLimitedChannel struct {
	inner   Channel
	limiter *ratelimit.Limiter
}

// NewRateLimitedChannel wraps inner with a rate limiter allowing at most max
// alerts per window for each unique alert key.
func NewRateLimitedChannel(inner Channel, max int, window time.Duration) *RateLimitedChannel {
	return &RateLimitedChannel{
		inner:   inner,
		limiter: ratelimit.New(max, window),
	}
}

// Send forwards the alert to the inner channel if within the rate limit.
// When the limit is exceeded it emits a throttled notice instead.
func (r *RateLimitedChannel) Send(a alerting.Alert) error {
	key := alertKey(a)
	if r.limiter.Allow(key) {
		return r.inner.Send(a)
	}
	notice := alerting.Alert{
		Level:     alerting.LevelWarn,
		Message:   fmt.Sprintf("[rate-limited] alert suppressed for key %q", key),
		Timestamp: a.Timestamp,
		Details:   a.Details,
	}
	return r.inner.Send(notice)
}

// Name returns a descriptive name for the channel.
func (r *RateLimitedChannel) Name() string {
	return fmt.Sprintf("rate-limited(%s)", r.inner.Name())
}

// alertKey builds a deduplication key from the alert's level and message.
func alertKey(a alerting.Alert) string {
	return fmt.Sprintf("%s:%s", a.Level, a.Message)
}
