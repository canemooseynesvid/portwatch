package reporter

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"portwatch/internal/snapshot"
)

// Reporter formats and writes port snapshot summaries to an output writer.
type Reporter struct {
	out io.Writer
}

// New creates a Reporter that writes to the given writer.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Reporter {
	if w == nil {
		w = os.Stdout
	}
	return &Reporter{out: w}
}

// PrintSnapshot writes a formatted table of all current port entries.
func (r *Reporter) PrintSnapshot(s *snapshot.Snapshot) {
	tw := tabwriter.NewWriter(r.out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "PROTOCOL\tLOCAL ADDRESS\tPORT\tPID\tSTATE")
	fmt.Fprintln(tw, "--------\t-------------\t----\t---\t-----")
	for _, e := range s.All() {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n",
			e.Protocol,
			e.LocalAddr,
			e.Port,
			e.PID,
			e.State,
		)
	}
	tw.Flush()
}

// PrintHistory writes a formatted list of recent port events from history.
func (r *Reporter) PrintHistory(h *snapshot.History) {
	events := h.Recent()
	if len(events) == 0 {
		fmt.Fprintln(r.out, "No recent events.")
		return
	}
	tw := tabwriter.NewWriter(r.out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "TIME\tEVENT\tPROTOCOL\tADDRESS\tPORT")
	fmt.Fprintln(tw, "----\t-----\t--------\t-------\t----")
	for _, ev := range events {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n",
			ev.At.Format(time.RFC3339),
			ev.Kind,
			ev.Entry.Protocol,
			ev.Entry.LocalAddr,
			ev.Entry.Port,
		)
	}
	tw.Flush()
}
