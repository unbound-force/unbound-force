package orchestration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkflowConfig_FileExists(t *testing.T) {
	dir := t.TempDir()

	content := `workflow:
  execution_modes:
    define: swarm
  spec_review: true
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadWorkflowConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkflowConfig() error: %v", err)
	}

	if cfg.Workflow.ExecutionModes["define"] != "swarm" {
		t.Errorf("ExecutionModes[define] = %q, want %q",
			cfg.Workflow.ExecutionModes["define"], "swarm")
	}
	if !cfg.Workflow.SpecReview {
		t.Error("SpecReview = false, want true")
	}
}

func TestLoadWorkflowConfig_FileMissing(t *testing.T) {
	dir := t.TempDir() // empty — no config.yaml

	cfg, err := LoadWorkflowConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkflowConfig() error: %v", err)
	}

	// Zero-value config: nil map, false spec_review
	if cfg.Workflow.ExecutionModes != nil {
		t.Errorf("ExecutionModes = %v, want nil", cfg.Workflow.ExecutionModes)
	}
	if cfg.Workflow.SpecReview {
		t.Error("SpecReview = true, want false")
	}
}

func TestLoadWorkflowConfig_Malformed(t *testing.T) {
	dir := t.TempDir()

	// Invalid YAML: tab indentation mixed with spaces, unclosed bracket
	content := `workflow: [invalid yaml {{`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadWorkflowConfig(dir)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoadWorkflowConfig_CommentedOut(t *testing.T) {
	dir := t.TempDir()

	// Scaffolded config with all values commented out.
	// Commented YAML = empty document = zero-value config.
	content := `# .unbound-force/config.yaml
# Workflow configuration for Unbound Force hero lifecycle.
# CLI flags (--define-mode, --spec-review) override these defaults.

# workflow:
#   execution_modes:
#     define: swarm
#   spec_review: false
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadWorkflowConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkflowConfig() error: %v", err)
	}

	if cfg.Workflow.ExecutionModes != nil {
		t.Errorf("ExecutionModes = %v, want nil (commented-out YAML)", cfg.Workflow.ExecutionModes)
	}
	if cfg.Workflow.SpecReview {
		t.Error("SpecReview = true, want false (commented-out YAML)")
	}
}
