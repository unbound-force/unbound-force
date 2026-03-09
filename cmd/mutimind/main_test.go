package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

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
