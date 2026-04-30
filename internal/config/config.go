package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// PortRule identifies a single port + protocol pair used in allow/deny lists.
type PortRule struct {
	Port     int    `toml:"port"`
	Protocol string `toml:"protocol"`
}

// Config holds all runtime configuration for portwatch.
type Config struct {
	PollInterval time.Duration `toml:"poll_interval"`
	LogLevel     string        `toml:"log_level"`
	AlertOnNew   bool          `toml:"alert_on_new"`
	AlertOnClose bool          `toml:"alert_on_close"`
	Privileged   bool          `toml:"alert_privileged"`
	AllowedPorts []PortRule    `toml:"allowed_ports"`
	DeniedPorts  []PortRule    `toml:"denied_ports"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		PollInterval: 5 * time.Second,
		LogLevel:     "info",
		AlertOnNew:   true,
		AlertOnClose: false,
		Privileged:   true,
		AllowedPorts: []PortRule{},
		DeniedPorts:  []PortRule{},
	}
}

// Load reads a TOML config file from path, merging it over the defaults.
func Load(path string) (*Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	dec := toml.NewDecoder(f)
	meta, err := dec.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}
	if len(meta.Undecoded()) > 0 {
		return nil, fmt.Errorf("config: unknown fields: %v", meta.Undecoded())
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.PollInterval < time.Second {
		return errors.New("config: poll_interval must be >= 1s")
	}
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.LogLevel] {
		return fmt.Errorf("config: unknown log_level %q", cfg.LogLevel)
	}
	for _, r := range append(cfg.AllowedPorts, cfg.DeniedPorts...) {
		if r.Port < 1 || r.Port > 65535 {
			return fmt.Errorf("config: port %d out of range", r.Port)
		}
		if r.Protocol != "tcp" && r.Protocol != "udp" {
			return fmt.Errorf("config: unknown protocol %q", r.Protocol)
		}
	}
	return nil
}
