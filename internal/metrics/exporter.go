package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

// Snapshot holds a point-in-time copy of all metric values.
type ExportSnapshot struct {
	Timestamp time.Time        `json:"timestamp"`
	Counters  map[string]int64 `json:"counters"`
}

// Exporter formats and writes metric snapshots to an output sink.
type Exporter struct {
	reg *Registry
	out io.Writer
}

// NewExporter returns an Exporter backed by reg that writes to out.
// If out is nil, os.Stdout is used.
func NewExporter(reg *Registry, out io.Writer) *Exporter {
	if out == nil {
		out = os.Stdout
	}
	return &Exporter{reg: reg, out: out}
}

// ExportJSON writes the current metric snapshot as a single JSON line.
func (e *Exporter) ExportJSON() error {
	snap := e.buildSnapshot()
	enc := json.NewEncoder(e.out)
	enc.SetIndent("", "  ")
	return enc.Encode(snap)
}

// ExportText writes the current metric snapshot as a human-readable table.
func (e *Exporter) ExportText() error {
	snap := e.buildSnapshot()

	w := tabwriter.NewWriter(e.out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "portwatch metrics — %s\n", snap.Timestamp.Format(time.RFC3339))
	fmt.Fprintln(w, "COUNTER\tVALUE")
	fmt.Fprintln(w, "-------\t-----")

	keys := make([]string, 0, len(snap.Counters))
	for k := range snap.Counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(w, "%s\t%d\n", k, snap.Counters[k])
	}
	return w.Flush()
}

func (e *Exporter) buildSnapshot() ExportSnapshot {
	return ExportSnapshot{
		Timestamp: time.Now().UTC(),
		Counters:  e.reg.Snapshot(),
	}
}
