package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds the portwatch daemon configuration.
type Config struct {
	// PollInterval is how often the scanner checks for port changes.
	PollInterval time.Duration `json:"poll_interval"`

	// AllowedPorts is a list of ports that are expected/allowed to be bound.
	// Bindings on these ports will be reported at Info level instead of Warn.
	AllowedPorts []uint16 `json:"allowed_ports"`

	// PrivilegedThreshold is the port number below which a binding is
	// considered privileged. Defaults to 1024.
	PrivilegedThreshold uint16 `json:"privileged_threshold"`

	// LogFile is an optional path to write alerts to. If empty, stdout is used.
	LogFile string `json:"log_file"`

	// Protocols lists which protocols to monitor ("tcp", "tcp6", "udp", "udp6").
	// If empty, all supported protocols are monitored.
	Protocols []string `json:"protocols"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		PollInterval:        5 * time.Second,
		AllowedPorts:        []uint16{},
		PrivilegedThreshold: 1024,
		LogFile:             "",
		Protocols:           []string{"tcp", "tcp6", "udp", "udp6"},
	}
}

// Load reads a JSON config file from path and merges it over the defaults.
func Load(path string) (*Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: validate: %w", err)
	}

	return cfg, nil
}

// Validate checks that the configuration values are sensible.
func (c *Config) Validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("poll_interval must be positive, got %v", c.PollInterval)
	}
	if c.PrivilegedThreshold == 0 {
		return fmt.Errorf("privileged_threshold must be > 0")
	}
	allowed := map[string]bool{"tcp": true, "tcp6": true, "udp": true, "udp6": true}
	for _, p := range c.Protocols {
		if !allowed[p] {
			return fmt.Errorf("unknown protocol %q", p)
		}
	}
	return nil
}

// IsAllowedPort reports whether port is in the AllowedPorts list.
func (c *Config) IsAllowedPort(port uint16) bool {
	for _, p := range c.AllowedPorts {
		if p == port {
			return true
		}
	}
	return false
}
