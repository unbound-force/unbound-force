package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

func TestCLI_Integration(t *testing.T) {
	// Setup a temporary directory for the backlog
	tempDir, err := os.MkdirTemp("", "mutimind-cli-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Override global variables for testing
	backlogDir = filepath.Join(tempDir, "backlog")
	artifactsDir = filepath.Join(tempDir, "artifacts")
	repo = backlog.NewRepository(backlogDir)

	// Helper to run commands
	runCmd := func(args ...string) string {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)
		err := rootCmd.Execute()
		if err != nil {
			t.Logf("Command failed: %v", err)
		}
		return buf.String()
	}

	// Test Add
	runCmd("add", "--title", "CLI Test", "--type", "story", "--priority", "P2", "--description", "CLI Description")

	// Test List Text
	out := runCmd("list")
	if out == "" {
		t.Errorf("Expected list output, got empty")
	}

	// Test List JSON
	outJSON := runCmd("list", "--format", "json")
	var items []backlog.Item
	if err := json.Unmarshal([]byte(outJSON), &items); err != nil {
		t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, outJSON)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 item in JSON output, got %d", len(items))
	}
	if items[0].Title != "CLI Test" {
		t.Errorf("Expected title 'CLI Test', got '%s'", items[0].Title)
	}

	// Test Update
	id := items[0].ID
	runCmd("update", id, "--status", "ready")

	// Test Show
	outShow := runCmd("show", id, "--format", "json")
	var item backlog.Item
	if err := json.Unmarshal([]byte(outShow), &item); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}
	if item.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", item.Status)
	}
}
