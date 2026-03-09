package artifacts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/artifacts"
	"github.com/unbound-force/unbound-force/internal/backlog"
)

func TestGenerateBacklogItemArtifact(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	item := &backlog.Item{
		ID:         "BI-123",
		Title:      "Test Item",
		Type:       "story",
		Priority:   "P2",
		Status:     "ready",
		CreatedAt:  now,
		ModifiedAt: now,
	}

	err := artifacts.GenerateBacklogItemArtifact(dir, item)
	if err != nil {
		t.Fatalf("Failed to generate artifact: %v", err)
	}

	expectedFile := filepath.Join(dir, "BI-123-backlog-item.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s was not created", expectedFile)
	}

	b, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read artifact: %v", err)
	}

	var env artifacts.Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if env.Hero != "muti-mind" {
		t.Errorf("Expected Hero 'muti-mind', got '%s'", env.Hero)
	}
	if env.ArtifactType != "backlog-item" {
		t.Errorf("Expected ArtifactType 'backlog-item', got '%s'", env.ArtifactType)
	}
	if env.Version != artifacts.Version {
		t.Errorf("Expected Version '%s', got '%s'", artifacts.Version, env.Version)
	}
	if env.SchemaVersion != "1.0.0" {
		t.Errorf("Expected SchemaVersion '1.0.0', got '%s'", env.SchemaVersion)
	}
	if env.Timestamp == "" {
		t.Errorf("Expected non-empty Timestamp")
	} else if _, err := time.Parse(time.RFC3339, env.Timestamp); err != nil {
		t.Errorf("Timestamp is not valid RFC3339: %v", err)
	}

	var payload backlog.Item
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if payload.ID != "BI-123" {
		t.Errorf("Expected payload ID 'BI-123', got '%s'", payload.ID)
	}
}

func TestGenerateAcceptanceDecision(t *testing.T) {
	dir := t.TempDir()
	dec := &artifacts.AcceptanceDecision{
		ItemID:         "BI-456",
		Decision:       "accept",
		Rationale:      "Looks good",
		CriteriaMet:    []string{"Test 1"},
		CriteriaFailed: []string{},
		GazeReportRef:  "report.json",
		DecidedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	err := artifacts.GenerateAcceptanceDecision(dir, dec)
	if err != nil {
		t.Fatalf("Failed to generate artifact: %v", err)
	}

	expectedFile := filepath.Join(dir, "BI-456-acceptance-decision.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s was not created", expectedFile)
	}

	b, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read artifact: %v", err)
	}

	var env artifacts.Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if env.ArtifactType != "acceptance-decision" {
		t.Errorf("Expected ArtifactType 'acceptance-decision', got '%s'", env.ArtifactType)
	}
	if env.Version != artifacts.Version {
		t.Errorf("Expected Version '%s', got '%s'", artifacts.Version, env.Version)
	}
	if env.SchemaVersion != "1.0.0" {
		t.Errorf("Expected SchemaVersion '1.0.0', got '%s'", env.SchemaVersion)
	}
	if env.Timestamp == "" {
		t.Errorf("Expected non-empty Timestamp")
	} else if _, err := time.Parse(time.RFC3339, env.Timestamp); err != nil {
		t.Errorf("Timestamp is not valid RFC3339: %v", err)
	}

	var payload artifacts.AcceptanceDecision
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if payload.Decision != "accept" {
		t.Errorf("Expected Decision 'accept', got '%s'", payload.Decision)
	}
}

func TestWriteArtifact_ErrorDir(t *testing.T) {
	// Creating a file then using it as a directory to force MkdirAll to fail
	dir := t.TempDir()
	fileAsDir := filepath.Join(dir, "file.txt")
	os.WriteFile(fileAsDir, []byte(""), 0644)

	err := artifacts.GenerateBacklogItemArtifact(fileAsDir, &backlog.Item{ID: "BI-999"})
	if err == nil {
		t.Errorf("Expected error when target dir is a file, got nil")
	}
}
