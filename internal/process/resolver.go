package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Info holds process metadata associated with a port binding.
type Info struct {
	PID  int
	Name string
	Exe  string
}

func (i Info) String() string {
	if i.PID == 0 {
		return "unknown"
	}
	return fmt.Sprintf("%s (pid %d)", i.Name, i.PID)
}

// Resolver looks up process information by inode number.
type Resolver struct {
	procRoot string
}

// New returns a Resolver that reads from the given procfs root (usually "/proc").
func New(procRoot string) *Resolver {
	return &Resolver{procRoot: procRoot}
}

// ByInode searches all process fd directories for a socket matching the given
// inode, then returns process metadata for the owning process.
func (r *Resolver) ByInode(inode uint64) (Info, error) {
	target := fmt.Sprintf("socket:[%d]", inode)

	entries, err := os.ReadDir(r.procRoot)
	if err != nil {
		return Info{}, fmt.Errorf("read procfs: %w", err)
	}

	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // skip non-numeric entries
		}

		fdDir := filepath.Join(r.procRoot, e.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				return r.infoForPID(pid)
			}
		}
	}

	return Info{}, nil
}

func (r *Resolver) infoForPID(pid int) (Info, error) {
	info := Info{PID: pid}

	exePath := filepath.Join(r.procRoot, strconv.Itoa(pid), "exe")
	if exe, err := os.Readlink(exePath); err == nil {
		info.Exe = exe
		info.Name = filepath.Base(exe)
	}

	commPath := filepath.Join(r.procRoot, strconv.Itoa(pid), "comm")
	if data, err := os.ReadFile(commPath); err == nil {
		info.Name = strings.TrimSpace(string(data))
	}

	return info, nil
}
