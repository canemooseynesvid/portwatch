package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Enricher augments port entries with process metadata using the resolver.
type Enricher struct {
	resolver *Resolver
}

// NewEnricher creates an Enricher backed by the given Resolver.
func NewEnricher(r *Resolver) *Enricher {
	return &Enricher{resolver: r}
}

// EnrichedInfo combines an inode with its resolved process information.
type EnrichedInfo struct {
	Inode  uint64
	Info   Info
	Found  bool
}

// Enrich resolves process info for the given inode.
func (e *Enricher) Enrich(inode uint64) EnrichedInfo {
	info, ok := e.resolver.ByInode(inode)
	return EnrichedInfo{
		Inode: inode,
		Info:  info,
		Found: ok,
	}
}

// EnrichMany resolves process info for a slice of inodes.
func (e *Enricher) EnrichMany(inodes []uint64) []EnrichedInfo {
	results := make([]EnrichedInfo, 0, len(inodes))
	for _, inode := range inodes {
		results = append(results, e.Enrich(inode))
	}
	return results
}

// FormatSummary returns a human-readable summary for an EnrichedInfo.
func FormatSummary(ei EnrichedInfo) string {
	if !ei.Found {
		return fmt.Sprintf("inode=%d (process not found)", ei.Inode)
	}
	return fmt.Sprintf("inode=%d pid=%d cmd=%s", ei.Inode, ei.Info.PID, ei.Info.Comm)
}

// ReadComm reads the comm name for a given PID from /proc.
func ReadComm(pid int) (string, error) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
