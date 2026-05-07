package process

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInfoString_WithPID(t *testing.T) {
	i := Info{PID: 42, Name: "nginx"}
	got := i.String()
	if got != "nginx (pid 42)" {
		t.Errorf("unexpected string: %q", got)
	}
}

func TestInfoString_ZeroPID(t *testing.T) {
	i := Info{}
	if i.String() != "unknown" {
		t.Errorf("expected 'unknown', got %q", i.String())
	}
}

// buildFakeProcFS creates a minimal /proc-like directory tree with a single
// process that owns a socket inode.
func buildFakeProcFS(t *testing.T, pid int, inode uint64, comm string) string {
	t.Helper()
	root := t.TempDir()

	pidDir := filepath.Join(root, fmt.Sprintf("%d", pid))
	fdDir := filepath.Join(pidDir, "fd")
	if err := os.MkdirAll(fdDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write comm file
	if err := os.WriteFile(filepath.Join(pidDir, "comm"), []byte(comm+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink that looks like a socket fd
	socketTarget := fmt.Sprintf("socket:[%d]", inode)
	if err := os.Symlink(socketTarget, filepath.Join(fdDir, "3")); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestByInode_FindsProcess(t *testing.T) {
	const pid = 1234
	const inode = uint64(99887766)
	const comm = "portwatch"

	root := buildFakeProcFS(t, pid, inode, comm)
	r := New(root)

	info, err := r.ByInode(inode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.PID != pid {
		t.Errorf("expected pid %d, got %d", pid, info.PID)
	}
	if info.Name != comm {
		t.Errorf("expected name %q, got %q", comm, info.Name)
	}
}

func TestByInode_NotFound(t *testing.T) {
	const pid = 5678
	const inode = uint64(11111)
	const otherInode = uint64(22222)

	root := buildFakeProcFS(t, pid, inode, "other")
	r := New(root)

	info, err := r.ByInode(otherInode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.PID != 0 {
		t.Errorf("expected zero PID for missing inode, got %d", info.PID)
	}
}

func TestByInode_BadProcRoot(t *testing.T) {
	r := New("/nonexistent/proc")
	_, err := r.ByInode(12345)
	if err == nil {
		t.Error("expected error for missing procfs root")
	}
}
