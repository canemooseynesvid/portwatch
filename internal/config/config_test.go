package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.PollInterval != 5*time.Second {
		t.Errorf("expected 5s poll interval, got %v", cfg.PollInterval)
	}
	if cfg.PrivilegedThreshold != 1024 {
		t.Errorf("expected privileged threshold 1024, got %d", cfg.PrivilegedThreshold)
	}
	if len(cfg.Protocols) != 4 {
		t.Errorf("expected 4 default protocols, got %d", len(cfg.Protocols))
	}
}

func TestLoad_ValidFile(t *testing.T) {
	data := map[string]interface{}{
		"poll_interval":        "2s",
		"allowed_ports":        []int{80, 443},
		"privileged_threshold": 1024,
		"protocols":            []string{"tcp", "tcp6"},
	}
	f := writeTempJSON(t, data)

	cfg, err := Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 2*time.Second {
		t.Errorf("expected 2s, got %v", cfg.PollInterval)
	}
	if len(cfg.AllowedPorts) != 2 {
		t.Errorf("expected 2 allowed ports, got %d", len(cfg.AllowedPorts))
	}
	if len(cfg.Protocols) != 2 {
		t.Errorf("expected 2 protocols, got %d", len(cfg.Protocols))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/portwatch.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_UnknownField(t *testing.T) {
	data := map[string]interface{}{"unknown_field": true}
	f := writeTempJSON(t, data)
	_, err := Load(f)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestValidate_BadPollInterval(t *testing.T) {
	cfg := Default()
	cfg.PollInterval = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for zero poll interval")
	}
}

func TestValidate_BadProtocol(t *testing.T) {
	cfg := Default()
	cfg.Protocols = []string{"tcp", "quic"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown protocol")
	}
}

func TestIsAllowedPort(t *testing.T) {
	cfg := Default()
	cfg.AllowedPorts = []uint16{80, 443, 8080}
	if !cfg.IsAllowedPort(80) {
		t.Error("expected 80 to be allowed")
	}
	if cfg.IsAllowedPort(9000) {
		t.Error("expected 9000 to not be allowed")
	}
}

// writeTempJSON writes v as JSON to a temp file and returns its path.
func writeTempJSON(t *testing.T, v interface{}) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if err := json.NewEncoder(f).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
	f.Close()
	return f.Name()
}
