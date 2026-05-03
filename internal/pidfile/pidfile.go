// Package pidfile provides PID file management for
// long-running background processes. It supports atomic
// writes, liveness checks via signal 0, and stale file
// cleanup. Extracted from internal/gateway to be shared
// by gateway, ollama-proxy, and future daemons.
//
// The PID file format is plain text with the PID on line 1
// and key=value metadata on subsequent lines.
package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PIDInfo represents the contents of a PID file. The file
// format is plain text with the PID on line 1 and
// key=value metadata on subsequent lines.
type PIDInfo struct {
	// PID is the process ID.
	PID int

	// Port is the local port the process listens on.
	Port int

	// Provider is the upstream provider name
	// ("anthropic", "vertex", "bedrock").
	Provider string

	// Started is the time the process was started.
	Started time.Time
}

// WritePID writes the PID file atomically (write to temp
// file, then rename) to prevent partial reads. Creates
// the parent directory if it does not exist.
//
// Design decision: Atomic write via temp+rename prevents
// a concurrent reader from seeing a half-written file.
func WritePID(path string, info PIDInfo) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create PID directory: %w", err)
	}

	content := fmt.Sprintf("%d\nport=%d\nprovider=%s\nstarted=%s\n",
		info.PID, info.Port, info.Provider, info.Started.Format(time.RFC3339))

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write PID temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename PID file: %w", err)
	}
	return nil
}

// ReadPID reads and parses the PID file. Returns an error
// if the file does not exist or contains an invalid PID.
// Unknown metadata keys are ignored for forward
// compatibility.
func ReadPID(path string) (*PIDInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read PID file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return nil, fmt.Errorf("PID file is empty")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid PID %q: %w", lines[0], err)
	}

	info := &PIDInfo{PID: pid}

	// Parse key=value metadata from subsequent lines.
	for _, line := range lines[1:] {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue // Skip malformed lines.
		}
		switch key {
		case "port":
			info.Port, _ = strconv.Atoi(value)
		case "provider":
			info.Provider = value
		case "started":
			info.Started, _ = time.Parse(time.RFC3339, value)
		}
		// Unknown keys are silently ignored for forward
		// compatibility.
	}

	return info, nil
}

// IsAlive checks whether a process with the given PID is
// still running. Uses the injected findProcess function
// and sends signal 0 (liveness check).
func IsAlive(pid int, findProcess func(int) (*os.Process, error)) bool {
	proc, err := findProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if the process exists and the caller
	// has permission to signal it. It does not actually
	// send a signal.
	err = proc.Signal(os.Signal(signalZero))
	return err == nil
}

// RemovePID removes the PID file. Returns nil if the file
// does not exist (idempotent).
func RemovePID(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove PID file: %w", err)
	}
	return nil
}

// CleanupStale reads the PID file, checks if the process
// is alive, and removes the PID file if the process is
// dead (stale). Returns nil if no PID file exists.
func CleanupStale(path string, findProcess func(int) (*os.Process, error)) error {
	info, err := ReadPID(path)
	if err != nil {
		// No PID file or unreadable — nothing to clean up.
		if os.IsNotExist(err) {
			return nil
		}
		// Malformed PID file — remove it.
		return RemovePID(path)
	}

	if IsAlive(info.PID, findProcess) {
		// Process is still running — not stale.
		return nil
	}

	// Process is dead — clean up the stale PID file.
	return RemovePID(path)
}
