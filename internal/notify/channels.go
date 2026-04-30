package notify

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// StdoutChannel writes notifications to stdout.
type StdoutChannel struct{}

func (s *StdoutChannel) Name() string { return "stdout" }

func (s *StdoutChannel) Send(subject, body string) error {
	_, err := fmt.Printf("[NOTIFY] %s\n%s\n", subject, body)
	return err
}

// WriterChannel writes notifications to an arbitrary io.Writer.
type WriterChannel struct {
	w    io.Writer
	name string
}

func NewWriterChannel(name string, w io.Writer) *WriterChannel {
	return &WriterChannel{name: name, w: w}
}

func (w *WriterChannel) Name() string { return w.name }

func (w *WriterChannel) Send(subject, body string) error {
	_, err := fmt.Fprintf(w.w, "[NOTIFY] %s\n%s\n", subject, body)
	return err
}

// ExecChannel runs an external command, passing subject and body as env vars.
type ExecChannel struct {
	command string
	args    []string
}

func NewExecChannel(command string, args ...string) *ExecChannel {
	return &ExecChannel{command: command, args: args}
}

func (e *ExecChannel) Name() string {
	return fmt.Sprintf("exec(%s)", e.command)
}

func (e *ExecChannel) Send(subject, body string) error {
	cmd := exec.Command(e.command, e.args...)
	cmd.Env = append(os.Environ(),
		"PORTWATCH_SUBJECT="+subject,
		"PORTWATCH_BODY="+strings.ReplaceAll(body, "\n", " "),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
