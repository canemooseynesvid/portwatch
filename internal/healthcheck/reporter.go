package healthcheck

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// Print writes a formatted health check report to w.
func Print(w io.Writer, results []Result) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tMESSAGE\tCHECKED AT")
	fmt.Fprintln(tw, "----\t------\t-------\t----------")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			r.Name,
			r.Status,
			r.Message,
			r.CheckedAt.Format("15:04:05"),
		)
	}
	_ = tw.Flush()

	overall := Overall(results)
	fmt.Fprintf(w, "\nOverall: %s\n", overall)
}

// PrintToStdout writes the health report to os.Stdout.
func PrintToStdout(results []Result) {
	Print(os.Stdout, results)
}
