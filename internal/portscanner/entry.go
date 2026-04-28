package portscanner

// PortEntry represents a single listening port discovered from /proc/net.
type PortEntry struct {
	Protocol  string
	LocalAddr string
	LocalPort uint16
	RemoteAddr string
	RemotePort uint16
	State     string
	PID       int
}

// Scanner holds the list of /proc/net paths to scan.
type Scanner struct {
	paths []string
}

// NewScanner creates a Scanner that reads from the given proc net paths.
// If paths is empty, default Linux paths are used.
func NewScanner(paths []string) *Scanner {
	if len(paths) == 0 {
		paths = []string{
			"/proc/net/tcp",
			"/proc/net/tcp6",
			"/proc/net/udp",
			"/proc/net/udp6",
		}
	}
	return &Scanner{paths: paths}
}

// Scan reads all configured proc net files and returns discovered port entries.
func (s *Scanner) Scan() ([]PortEntry, error) {
	var all []PortEntry
	for _, path := range s.paths {
		entries, err := parseProcNet(path)
		if err != nil {
			// Non-fatal: file may not exist on all systems.
			continue
		}
		proto := protocolFromPath(path)
		for i := range entries {
			entries[i].Protocol = proto
		}
		all = append(all, entries...)
	}
	return all, nil
}
