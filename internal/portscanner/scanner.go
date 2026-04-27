package portscanner

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PortEntry represents a single bound port with its metadata.
type PortEntry struct {
	Protocol string
	LocalAddr string
	Port     int
	PID      int
	State    string
}

// Scanner reads active port bindings from the system.
type Scanner struct {
	procNetPaths []string
}

// NewScanner creates a Scanner with default /proc paths.
func NewScanner() *Scanner {
	return &Scanner{
		procNetPaths: []string{"/proc/net/tcp", "/proc/net/tcp6", "/proc/net/udp", "/proc/net/udp6"},
	}
}

// Scan returns all currently bound port entries.
func (s *Scanner) Scan() ([]PortEntry, error) {
	var entries []PortEntry
	for _, path := range s.procNetPaths {
		protocol := protocolFromPath(path)
		results, err := parseProcNet(path, protocol)
		if err != nil {
			// Skip unavailable files (e.g., no IPv6 support)
			continue
		}
		entries = append(entries, results...)
	}
	return entries, nil
}

func protocolFromPath(path string) string {
	if strings.Contains(path, "udp") {
		return "udp"
	}
	return "tcp"
}

func parseProcNet(path, protocol string) ([]PortEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []PortEntry
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum == 1 {
			continue // skip header
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		localAddr, port, err := parseHexAddr(fields[1])
		if err != nil {
			continue
		}
		state := fields[3]
		entries = append(entries, PortEntry{
			Protocol:  protocol,
			LocalAddr: localAddr,
			Port:     port,
			State:    state,
		})
	}
	return entries, scanner.Err()
}

func parseHexAddr(hexAddr string) (string, int, error) {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address: %s", hexAddr)
	}
	portVal, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return "", 0, err
	}
	return parts[0], int(portVal), nil
}
