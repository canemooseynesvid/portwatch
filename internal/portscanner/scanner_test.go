package portscanner

import (
	"os"
	"testing"
)

func TestParseHexAddr_Valid(t *testing.T) {
	addr, port, err := parseHexAddr("0100007F:1F90")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 {
		t.Errorf("expected port 8080, got %d", port)
	}
	if addr != "0100007F" {
		t.Errorf("expected addr 0100007F, got %s", addr)
	}
}

func TestParseHexAddr_Invalid(t *testing.T) {
	_, _, err := parseHexAddr("baddata")
	if err == nil {
		t.Error("expected error for invalid address, got nil")
	}
}

func TestProtocolFromPath(t *testing.T) {
	tests := []struct {
		path     string
		want     string
	}{
		{"/proc/net/tcp", "tcp"},
		{"/proc/net/tcp6", "tcp"},
		{"/proc/net/udp", "udp"},
		{"/proc/net/udp6", "udp"},
	}
	for _, tt := range tests {
		got := protocolFromPath(tt.path)
		if got != tt.want {
			t.Errorf("protocolFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestParseProcNet_MissingFile(t *testing.T) {
	_, err := parseProcNet("/nonexistent/path", "tcp")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParseProcNet_ValidFile(t *testing.T) {
	content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1 0000000000000000 100 0 0 10 0
`
	tmpFile, err := os.CreateTemp("", "proc_net_tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	entries, err := parseProcNet(tmpFile.Name(), "tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", entries[0].Port)
	}
	if entries[0].Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %s", entries[0].Protocol)
	}
}
