package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWritePID_ReadPID_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".uf", "gateway.pid")

	started := time.Date(2026, 4, 20, 14, 30, 0, 0, time.UTC)
	info := PIDInfo{
		PID:      42195,
		Port:     53147,
		Provider: "vertex",
		Started:  started,
	}

	if err := WritePID(path, info); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	got, err := ReadPID(path)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}

	if got.PID != info.PID {
		t.Errorf("PID: got %d, want %d", got.PID, info.PID)
	}
	if got.Port != info.Port {
		t.Errorf("Port: got %d, want %d", got.Port, info.Port)
	}
	if got.Provider != info.Provider {
		t.Errorf("Provider: got %q, want %q", got.Provider, info.Provider)
	}
	if !got.Started.Equal(info.Started) {
		t.Errorf("Started: got %v, want %v", got.Started, info.Started)
	}
}

func TestReadPID_MalformedFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "non-numeric PID",
			content: "not-a-number\nport=53147\n",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "missing metadata",
			content: "12345\n",
			wantErr: false, // PID is valid, metadata is optional
		},
		{
			name:    "valid with unknown keys",
			content: "12345\nport=53147\nunknown=value\n",
			wantErr: false, // Unknown keys are ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "gateway.pid")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			info, err := ReadPID(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.PID != 12345 {
				t.Errorf("PID: got %d, want 12345", info.PID)
			}
		})
	}
}

func TestReadPID_FileNotFound(t *testing.T) {
	_, err := ReadPID("/nonexistent/path/gateway.pid")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestIsAlive_ProcessFound(t *testing.T) {
	findProcess := func(pid int) (*os.Process, error) {
		// Return a mock process that accepts signal 0.
		return os.FindProcess(os.Getpid())
	}

	// Use current process PID — it's definitely alive.
	alive := IsAlive(os.Getpid(), findProcess)
	if !alive {
		t.Error("expected alive=true for current process")
	}
}

func TestIsAlive_ProcessNotFound(t *testing.T) {
	findProcess := func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	alive := IsAlive(99999, findProcess)
	if alive {
		t.Error("expected alive=false when process not found")
	}
}

func TestRemovePID_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.pid")
	if err := os.WriteFile(path, []byte("12345\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemovePID(path); err != nil {
		t.Fatalf("RemovePID failed: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestRemovePID_NonExistent(t *testing.T) {
	// Removing a non-existent file should not error (idempotent).
	err := RemovePID("/nonexistent/path/gateway.pid")
	if err != nil {
		t.Fatalf("expected nil error for non-existent file, got: %v", err)
	}
}

func TestCleanupStale_DeadProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.pid")
	content := "99999\nport=53147\nprovider=anthropic\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Process is dead.
	findProcess := func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	if err := CleanupStale(path, findProcess); err != nil {
		t.Fatalf("CleanupStale failed: %v", err)
	}

	// PID file should be removed.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected stale PID file to be removed")
	}
}

func TestCleanupStale_AliveProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.pid")
	content := fmt.Sprintf("%d\nport=53147\nprovider=anthropic\n", os.Getpid())
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Process is alive (current process).
	findProcess := func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	if err := CleanupStale(path, findProcess); err != nil {
		t.Fatalf("CleanupStale failed: %v", err)
	}

	// PID file should still exist.
	if _, err := os.Stat(path); err != nil {
		t.Error("expected PID file to be preserved for alive process")
	}
}

func TestCleanupStale_NoPIDFile(t *testing.T) {
	err := CleanupStale("/nonexistent/path/gateway.pid", os.FindProcess)
	if err != nil {
		t.Fatalf("expected nil error when no PID file, got: %v", err)
	}
}

func TestWritePID_NonExistentDirectory(t *testing.T) {
	// WritePID should create the directory if it doesn't
	// exist (via MkdirAll).
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "dir", "gateway.pid")

	info := PIDInfo{
		PID:      12345,
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}

	if err := WritePID(path, info); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	// Verify the file was created.
	got, err := ReadPID(path)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	if got.PID != 12345 {
		t.Errorf("PID: got %d, want 12345", got.PID)
	}
}

func TestWritePID_ReadOnlyDirectory(t *testing.T) {
	// Test that WritePID returns an error when the
	// directory cannot be created.
	path := "/proc/nonexistent/gateway.pid"
	info := PIDInfo{PID: 1, Port: 1, Provider: "test"}

	err := WritePID(path, info)
	if err == nil {
		t.Fatal("expected error writing to read-only path")
	}
}
