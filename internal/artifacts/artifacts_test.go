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

func TestFindArtifacts_MatchesType(t *testing.T) {
	dir := t.TempDir()
	// Write 2 artifacts with different types
	if err := artifacts.WriteArtifact(dir, "gaze", "quality-report", "r1", map[string]string{"score": "A"}); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.WriteArtifact(dir, "divisor", "review-verdict", "v1", map[string]string{"decision": "approve"}); err != nil {
		t.Fatal(err)
	}

	paths, err := artifacts.FindArtifacts(dir, "quality-report")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 quality-report, got %d", len(paths))
	}
}

func TestFindArtifacts_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	paths, err := artifacts.FindArtifacts(dir, "quality-report")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 artifacts in empty dir, got %d", len(paths))
	}
}

func TestFindArtifacts_SkipsNonJSON(t *testing.T) {
	dir := t.TempDir()
	// Write a non-JSON file
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# hello"), 0644)
	// Write a valid artifact
	artifacts.WriteArtifact(dir, "gaze", "quality-report", "r1", map[string]string{})

	paths, err := artifacts.FindArtifacts(dir, "quality-report")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 artifact (skipping .md), got %d", len(paths))
	}
}

func TestReadEnvelope_ValidFile(t *testing.T) {
	dir := t.TempDir()
	artifacts.WriteArtifact(dir, "mx-f", "metrics-snapshot", "s1", map[string]float64{"velocity": 8.2})

	// Find the written file
	paths, _ := artifacts.FindArtifacts(dir, "metrics-snapshot")
	if len(paths) == 0 {
		t.Fatal("no artifact found")
	}

	env, err := artifacts.ReadEnvelope(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if env.Hero != "mx-f" {
		t.Errorf("Hero = %q", env.Hero)
	}
	if env.ArtifactType != "metrics-snapshot" {
		t.Errorf("Type = %q", env.ArtifactType)
	}
}

func TestReadEnvelope_MalformedFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{invalid json"), 0644)
	_, err := artifacts.ReadEnvelope(filepath.Join(dir, "bad.json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestWriteArtifact_WithContext(t *testing.T) {
	dir := t.TempDir()
	ctx := &artifacts.ArtifactContext{
		Branch:        "feat/health-check",
		Commit:        "abc123",
		BacklogItemID: "BI-042",
		CorrelationID: "corr-001",
		WorkflowID:    "wf-feat-health-check-20260320",
	}

	err := artifacts.WriteArtifactWithContext(dir, "gaze", "quality-report", "qr-001", map[string]string{"score": "A"}, ctx)
	if err != nil {
		t.Fatalf("WriteArtifactWithContext failed: %v", err)
	}

	paths, err := artifacts.FindArtifacts(dir, "quality-report")
	if err != nil {
		t.Fatalf("FindArtifacts failed: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(paths))
	}

	env, err := artifacts.ReadEnvelope(paths[0])
	if err != nil {
		t.Fatalf("ReadEnvelope failed: %v", err)
	}

	if env.Context == nil {
		t.Fatal("expected non-nil Context")
	}

	var readCtx artifacts.ArtifactContext
	if err := json.Unmarshal(env.Context, &readCtx); err != nil {
		t.Fatalf("unmarshal context: %v", err)
	}

	if readCtx.Branch != "feat/health-check" {
		t.Errorf("Branch = %q, want %q", readCtx.Branch, "feat/health-check")
	}
	if readCtx.WorkflowID != "wf-feat-health-check-20260320" {
		t.Errorf("WorkflowID = %q, want %q", readCtx.WorkflowID, "wf-feat-health-check-20260320")
	}
	if readCtx.BacklogItemID != "BI-042" {
		t.Errorf("BacklogItemID = %q, want %q", readCtx.BacklogItemID, "BI-042")
	}
	if readCtx.Commit != "abc123" {
		t.Errorf("Commit = %q, want %q", readCtx.Commit, "abc123")
	}
	if readCtx.CorrelationID != "corr-001" {
		t.Errorf("CorrelationID = %q, want %q", readCtx.CorrelationID, "corr-001")
	}
}

func TestWriteArtifact_WithNilContext(t *testing.T) {
	dir := t.TempDir()

	err := artifacts.WriteArtifactWithContext(dir, "gaze", "quality-report", "qr-002", map[string]string{"score": "B"}, nil)
	if err != nil {
		t.Fatalf("WriteArtifactWithContext with nil context failed: %v", err)
	}

	paths, err := artifacts.FindArtifacts(dir, "quality-report")
	if err != nil {
		t.Fatalf("FindArtifacts failed: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(paths))
	}

	env, err := artifacts.ReadEnvelope(paths[0])
	if err != nil {
		t.Fatalf("ReadEnvelope failed: %v", err)
	}

	// Context should be empty/nil when no context provided
	if len(env.Context) > 0 {
		t.Errorf("expected empty Context for nil input, got %s", string(env.Context))
	}
}

func TestFindArtifactsByHero_FiltersCorrectly(t *testing.T) {
	dir := t.TempDir()

	// Write artifacts from different heroes
	if err := artifacts.WriteArtifact(dir, "gaze", "quality-report", "qr-1", map[string]string{"score": "A"}); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.WriteArtifact(dir, "divisor", "quality-report", "qr-2", map[string]string{"score": "B"}); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.WriteArtifact(dir, "gaze", "quality-report", "qr-3", map[string]string{"score": "C"}); err != nil {
		t.Fatal(err)
	}

	paths, err := artifacts.FindArtifactsByHero(dir, "quality-report", "gaze")
	if err != nil {
		t.Fatalf("FindArtifactsByHero failed: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 gaze quality-reports, got %d", len(paths))
	}

	// Verify divisor is excluded
	pathsDivisor, err := artifacts.FindArtifactsByHero(dir, "quality-report", "divisor")
	if err != nil {
		t.Fatalf("FindArtifactsByHero failed: %v", err)
	}
	if len(pathsDivisor) != 1 {
		t.Errorf("expected 1 divisor quality-report, got %d", len(pathsDivisor))
	}
}

func TestFindArtifactsSince_FiltersByTime(t *testing.T) {
	dir := t.TempDir()

	// Write two artifacts (both will have current timestamp)
	if err := artifacts.WriteArtifact(dir, "gaze", "quality-report", "qr-old", map[string]string{}); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.WriteArtifact(dir, "gaze", "quality-report", "qr-new", map[string]string{}); err != nil {
		t.Fatal(err)
	}

	// Since a time in the past should return all
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	paths, err := artifacts.FindArtifactsSince(dir, "quality-report", past)
	if err != nil {
		t.Fatalf("FindArtifactsSince failed: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 artifacts since 2020, got %d", len(paths))
	}

	// Since a time in the future should return none
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	pathsFuture, err := artifacts.FindArtifactsSince(dir, "quality-report", future)
	if err != nil {
		t.Fatalf("FindArtifactsSince failed: %v", err)
	}
	if len(pathsFuture) != 0 {
		t.Errorf("expected 0 artifacts since 2030, got %d", len(pathsFuture))
	}
}

func TestCheckSchemaVersion_Compatible(t *testing.T) {
	env := &artifacts.Envelope{SchemaVersion: "1.0.0"}

	// Exact match
	compat, warn := artifacts.CheckSchemaVersion(env, "1.0.0")
	if !compat {
		t.Error("expected compatible for exact match")
	}
	if warn != "" {
		t.Errorf("expected no warning for exact match, got %q", warn)
	}

	// Minor version differs
	compat, warn = artifacts.CheckSchemaVersion(env, "1.1.0")
	if !compat {
		t.Error("expected compatible for minor version difference")
	}
	if warn == "" {
		t.Error("expected warning for minor version difference")
	}
}

func TestCheckSchemaVersion_MajorMismatch(t *testing.T) {
	env := &artifacts.Envelope{SchemaVersion: "2.0.0"}

	compat, warn := artifacts.CheckSchemaVersion(env, "1.0.0")
	if compat {
		t.Error("expected incompatible for major version mismatch")
	}
	if warn == "" {
		t.Error("expected warning for major version mismatch")
	}
}

func TestWriteArtifact_CustomHero(t *testing.T) {
	dir := t.TempDir()
	err := artifacts.WriteArtifact(dir, "mx-f", "metrics-snapshot", "snap-001", map[string]string{"v": "1"})
	if err != nil {
		t.Fatal(err)
	}

	paths, _ := artifacts.FindArtifacts(dir, "metrics-snapshot")
	if len(paths) != 1 {
		t.Fatal("expected 1 artifact")
	}

	env, _ := artifacts.ReadEnvelope(paths[0])
	if env.Hero != "mx-f" {
		t.Errorf("Hero = %q, want mx-f", env.Hero)
	}
}
