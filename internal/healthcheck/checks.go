package healthcheck

import (
	"fmt"
	"os"
	"time"
)

// FuncChecker wraps a plain function as a Checker.
type FuncChecker struct {
	name string
	fn   func() Result
}

// NewFuncChecker creates a Checker from a name and function.
func NewFuncChecker(name string, fn func() Result) Checker {
	return &FuncChecker{name: name, fn: fn}
}

func (f *FuncChecker) Name() string  { return f.name }
func (f *FuncChecker) Check() Result { return f.fn() }

// ProcFSChecker verifies that /proc/net is accessible.
type ProcFSChecker struct{}

func NewProcFSChecker() Checker { return &ProcFSChecker{} }

func (p *ProcFSChecker) Name() string { return "procfs" }

func (p *ProcFSChecker) Check() Result {
	_, err := os.Stat("/proc/net")
	if err != nil {
		return Result{
			Name:    p.Name(),
			Status:  StatusFailed,
			Message: fmt.Sprintf("/proc/net not accessible: %v", err),
		}
	}
	return Result{Name: p.Name(), Status: StatusOK, Message: "/proc/net accessible"}
}

// UptimeChecker reports how long the process has been running.
type UptimeChecker struct {
	start time.Time
}

func NewUptimeChecker(start time.Time) Checker {
	return &UptimeChecker{start: start}
}

func (u *UptimeChecker) Name() string { return "uptime" }

func (u *UptimeChecker) Check() Result {
	uptime := time.Since(u.start).Round(time.Second)
	return Result{
		Name:    u.Name(),
		Status:  StatusOK,
		Message: fmt.Sprintf("running for %s", uptime),
	}
}
