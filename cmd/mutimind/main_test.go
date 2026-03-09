package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

// stubGHRunner satisfies sync.GHRunner for CLI-level tests.
type stubGHRunner struct {
	out []byte
	err error
}

func (s *stubGHRunner) Run(args ...string) ([]byte, error) {
	return s.out, s.err
}

func TestCLI_Add_CreatesItem(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", "--title", "CLI Test", "--type", "story", "--priority", "P2", "--description", "CLI Description", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	repo := backlog.NewRepository(backlogDir)
	items, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Title != "CLI Test" {
		t.Errorf("Expected title 'CLI Test', got '%s'", items[0].Title)
	}
	if items[0].Type != "story" {
		t.Errorf("Expected type 'story', got '%s'", items[0].Type)
	}
	if items[0].Priority != "P2" {
		t.Errorf("Expected priority 'P2', got '%s'", items[0].Priority)
	}
}

func TestCLI_List_OutputsItems(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "CLI Test", Type: "story", Priority: "P2", Status: "draft"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"list", "--format", "json", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	var items []backlog.Item
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item in JSON output, got %d", len(items))
	}
	if items[0].Title != "CLI Test" {
		t.Errorf("Expected title 'CLI Test', got '%s'", items[0].Title)
	}
}

func TestCLI_Update_ModifiesItem(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "CLI Test", Type: "story", Priority: "P2", Status: "draft"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"update", "BI-001", "--status", "ready", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Update command failed: %v", err)
	}

	item, err := repo.Get("BI-001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if item.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", item.Status)
	}
}

func TestCLI_Show_OutputsItemDetails(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "CLI Test", Type: "story", Priority: "P2", Status: "ready"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"show", "BI-001", "--format", "json", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Show command failed: %v", err)
	}

	var item backlog.Item
	if err := json.Unmarshal(buf.Bytes(), &item); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}
	if item.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", item.Status)
	}
	if item.ID != "BI-001" {
		t.Errorf("Expected ID 'BI-001', got '%s'", item.ID)
	}
}

func TestCLI_Decide_CreatesArtifact(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "CLI Test"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"decide", "--item", "BI-001", "--decision", "accept", "--rationale", "looks good", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Decide command failed: %v", err)
	}

	artifactPath := filepath.Join(artifactsDir, "BI-001-acceptance-decision.json")
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		t.Fatalf("Artifact file not created")
	}

	b, _ := os.ReadFile(artifactPath)
	var env struct {
		Payload struct {
			Decision string `json:"decision"`
		} `json:"payload"`
	}
	_ = json.Unmarshal(b, &env)
	if env.Payload.Decision != "accept" {
		t.Errorf("Expected decision 'accept', got '%s'", env.Payload.Decision)
	}
}

func TestCLI_Init_CreatesDirectories(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	if _, err := os.Stat(backlogDir); err != nil {
		t.Errorf("Expected backlog dir to exist: %v", err)
	}
	if _, err := os.Stat(artifactsDir); err != nil {
		t.Errorf("Expected artifacts dir to exist: %v", err)
	}
}

func TestCLI_Show_TextOutput(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Show Test", Type: "story", Priority: "P1", Status: "ready"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// default format is text
	rootCmd.SetArgs([]string{"show", "BI-001", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Show command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "BI-001") {
		t.Errorf("Expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "Show Test") {
		t.Errorf("Expected title in output, got: %s", out)
	}
}

func TestCLI_List_FilterByStatus(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Draft Item", Type: "story", Priority: "P2", Status: "draft"})
	_ = repo.Save(&backlog.Item{ID: "BI-002", Title: "Ready Item", Type: "story", Priority: "P1", Status: "ready"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"list", "--status", "ready", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "Draft Item") {
		t.Errorf("Expected draft item to be filtered out, got: %s", out)
	}
	if !strings.Contains(out, "Ready Item") {
		t.Errorf("Expected ready item in output, got: %s", out)
	}
}

func TestCLI_List_TextOutput(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Text List Item", Type: "task", Priority: "P3", Status: "draft"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// default format is text (no --format flag)
	rootCmd.SetArgs([]string{"list", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "BI-001") {
		t.Errorf("Expected item ID in text output, got: %s", out)
	}
}

func TestCLI_SyncPush_WithStubRunner(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Sync Push Item"})

	stub := &stubGHRunner{out: []byte("https://github.com/org/repo/issues/10\n")}

	rootCmd := newRootCmdWithParams(&AppParams{GHRunner: stub})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sync-push", "BI-001", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync-push failed: %v", err)
	}

	item, _ := repo.Get("BI-001")
	if item.GitHubIssueNumber == nil || *item.GitHubIssueNumber != 10 {
		t.Errorf("Expected issue #10, got %v", item.GitHubIssueNumber)
	}
}

func TestCLI_SyncPull_WithStubRunner(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	stub := &stubGHRunner{
		out: []byte(`[{"number":5,"title":"Pulled Issue","body":"body","state":"OPEN","updatedAt":"2023-01-01T00:00:00Z"}]`),
	}

	rootCmd := newRootCmdWithParams(&AppParams{GHRunner: stub})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sync-pull", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync-pull failed: %v", err)
	}

	repo := backlog.NewRepository(backlogDir)
	items, _ := repo.List()
	if len(items) != 1 {
		t.Fatalf("Expected 1 imported item, got %d", len(items))
	}
}

func TestCLI_SyncStatus_WithStubRunner(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	num := 7
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Synced", GitHubIssueNumber: &num})

	stub := &stubGHRunner{}
	rootCmd := newRootCmdWithParams(&AppParams{GHRunner: stub})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sync-status", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync-status failed: %v", err)
	}

	if !strings.Contains(buf.String(), "synced") {
		t.Errorf("Expected 'synced' in output, got: %s", buf.String())
	}
}

func TestCLI_Sync_WithStubRunner(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	stub := &stubGHRunner{out: []byte(`[]`)}
	rootCmd := newRootCmdWithParams(&AppParams{GHRunner: stub})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sync", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Pulling") {
		t.Errorf("Expected 'Pulling' in output, got: %s", out)
	}
}

func TestCLI_SyncProject_WithStubRunner(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	stub := &stubGHRunner{}
	rootCmd := newRootCmdWithParams(&AppParams{GHRunner: stub})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sync-project", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync-project failed: %v", err)
	}
}

func TestCLI_Add_MissingRequiredFlags(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing required flags, got nil")
	}
}

func TestCLI_Decide_InvalidDecision(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"decide", "--item", "BI-001", "--decision", "maybe", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid decision, got nil")
	}
}

func TestCLI_GenerateArtifact_CreatesArtifact(t *testing.T) {
	tempDir := t.TempDir()
	backlogDir := filepath.Join(tempDir, "backlog")
	artifactsDir := filepath.Join(tempDir, "artifacts")

	repo := backlog.NewRepository(backlogDir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "CLI Test"})

	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"generate-artifact", "BI-001", "--backlog-dir", backlogDir, "--artifacts-dir", artifactsDir})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Generate-artifact command failed: %v", err)
	}

	artifactPath := filepath.Join(artifactsDir, "BI-001-backlog-item.json")
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		t.Fatalf("Artifact file not created")
	}

	b, _ := os.ReadFile(artifactPath)
	var env struct {
		ArtifactType string `json:"artifact_type"`
	}
	_ = json.Unmarshal(b, &env)
	if env.ArtifactType != "backlog-item" {
		t.Errorf("Expected artifact type 'backlog-item', got '%s'", env.ArtifactType)
	}
}
