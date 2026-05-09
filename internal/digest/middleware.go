package digest

import (
	"context"
	"log/slog"
)

// ScanResult carries the entries produced by a single scanner poll.
type ScanResult struct {
	Entries []Entry
}

// Scanner is the interface satisfied by the port scanner.
type Scanner interface {
	Scan(ctx context.Context) ([]Entry, error)
}

// ChangeFilter wraps a Scanner and skips downstream processing when the
// snapshot digest has not changed since the previous scan. This reduces
// unnecessary alerting and diff work during quiet periods.
type ChangeFilter struct {
	scanner Scanner
	tracker *Tracker
	log     *slog.Logger
}

// NewChangeFilter returns a ChangeFilter backed by the given scanner.
func NewChangeFilter(s Scanner, log *slog.Logger) *ChangeFilter {
	if log == nil {
		log = slog.Default()
	}
	return &ChangeFilter{
		scanner: s,
		tracker: New(),
		log:     log,
	}
}

// Scan delegates to the underlying scanner and returns (result, true) when
// the digest has changed, or (zero, false) when the snapshot is identical to
// the previous poll. Errors from the underlying scanner are always returned.
func (f *ChangeFilter) Scan(ctx context.Context) (ScanResult, bool, error) {
	entries, err := f.scanner.Scan(ctx)
	if err != nil {
		return ScanResult{}, false, err
	}

	changed, d := f.tracker.Update(entries)
	if !changed {
		f.log.Debug("digest unchanged, skipping processing", "digest", d)
		return ScanResult{}, false, nil
	}

	f.log.Debug("digest changed", "digest", d, "entries", len(entries))
	return ScanResult{Entries: entries}, true, nil
}
