package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/config"
)

// --- runConfigInit tests ---

func TestRunConfigInit_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := runConfigInit(configInitParams{
		targetDir: dir,
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("runConfigInit error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created") {
		t.Errorf("expected 'Created' in output, got: %s", output)
	}
	if !strings.Contains(output, "commented out") {
		t.Errorf("expected guidance about commented values, got: %s", output)
	}

	// Verify file was actually created.
	configPath := filepath.Join(dir, ".uf", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestRunConfigInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run creates.
	if err := runConfigInit(configInitParams{targetDir: dir, stdout: &buf}); err != nil {
		t.Fatalf("first run error = %v", err)
	}

	// Second run reports up to date.
	buf.Reset()
	if err := runConfigInit(configInitParams{targetDir: dir, stdout: &buf}); err != nil {
		t.Fatalf("second run error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "up to date") {
		t.Errorf("expected 'up to date' on idempotent run, got: %s", output)
	}
}

func TestRunConfigInit_UpdateAddsSections(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a config missing the gateway section.
	tmpl := config.Template()
	lines := strings.Split(tmpl, "\n")
	var filtered []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, "─── Gateway") {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(line, "# ───") {
			skip = false
		}
		if !skip {
			filtered = append(filtered, line)
		}
	}
	if err := os.WriteFile(
		filepath.Join(ufDir, "config.yaml"),
		[]byte(strings.Join(filtered, "\n")), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runConfigInit(configInitParams{targetDir: dir, stdout: &buf}); err != nil {
		t.Fatalf("runConfigInit error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", output)
	}
	if !strings.Contains(output, "gateway") {
		t.Errorf("expected 'gateway' in added sections, got: %s", output)
	}
}

// --- runConfigShow tests ---

func TestRunConfigShow_YAML(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := runConfigShow(configShowParams{
		targetDir: dir,
		format:    "text",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("runConfigShow error = %v", err)
	}

	output := buf.String()
	// Should contain default values in YAML format.
	if !strings.Contains(output, "granite-embedding:30m") {
		t.Errorf("expected default embedding model in YAML output, got: %s", output)
	}
	if !strings.Contains(output, "package_manager") {
		t.Errorf("expected 'package_manager' in YAML output, got: %s", output)
	}
}

func TestRunConfigShow_JSON(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := runConfigShow(configShowParams{
		targetDir: dir,
		format:    "json",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("runConfigShow error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"model"`) {
		t.Errorf("expected JSON key 'model' in output, got: %s", output)
	}
	if !strings.Contains(output, `"granite-embedding:30m"`) {
		t.Errorf("expected default embedding model in JSON output, got: %s", output)
	}
}

// --- runConfigValidate tests ---

func TestRunConfigValidate_NoFile(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := runConfigValidate(configValidateParams{
		targetDir: dir,
		format:    "text",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("runConfigValidate error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No config file found") {
		t.Errorf("expected 'No config file found', got: %s", output)
	}
	if !strings.Contains(output, "valid") {
		t.Errorf("expected 'valid' in output, got: %s", output)
	}
}

func TestRunConfigValidate_ValidFile(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte("setup:\n  package_manager: homebrew\ngateway:\n  port: 8080\n")
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runConfigValidate(configValidateParams{
		targetDir: dir,
		format:    "text",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("runConfigValidate error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "All checks passed") {
		t.Errorf("expected 'All checks passed', got: %s", output)
	}
}

func TestRunConfigValidate_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Invalid YAML: tab characters at start are not valid.
	cfgData := []byte("setup:\n\t\tbad: value\n")
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runConfigValidate(configValidateParams{
		targetDir: dir,
		format:    "text",
		stdout:    &buf,
	})
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}

	output := buf.String()
	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected 'FAIL' in output, got: %s", output)
	}
}

func TestRunConfigValidate_InvalidFieldValues(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte("setup:\n  package_manager: invalid_pm\ngateway:\n  provider: badprovider\n")
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runConfigValidate(configValidateParams{
		targetDir: dir,
		format:    "text",
		stdout:    &buf,
	})
	if err == nil {
		t.Fatal("expected error for invalid field values")
	}

	output := buf.String()
	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected 'FAIL' in output, got: %s", output)
	}
	if !strings.Contains(output, "package_manager") {
		t.Errorf("expected package_manager error, got: %s", output)
	}
}

// --- Field validator tests ---

func TestValidateSetup_ValidValues(t *testing.T) {
	errs := config.Validate(config.Config{Setup: config.SetupConfig{
		PackageManager: "homebrew",
		Tools: map[string]config.ToolConfig{
			"gaze": {Method: "rpm"},
			"node": {Method: "fnm", Version: "22"},
		},
	}})
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateSetup_InvalidPackageManager(t *testing.T) {
	errs := config.Validate(config.Config{Setup: config.SetupConfig{
		PackageManager: "invalid",
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateSetup_InvalidToolMethod(t *testing.T) {
	errs := config.Validate(config.Config{Setup: config.SetupConfig{
		Tools: map[string]config.ToolConfig{
			"gaze": {Method: "invalid"},
		},
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateSandbox_ValidValues(t *testing.T) {
	errs := config.Validate(config.Config{Sandbox: config.SandboxConfig{
		Runtime: "podman",
		Backend: "che",
		Mode:    "direct",
	}})
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateSandbox_InvalidRuntime(t *testing.T) {
	errs := config.Validate(config.Config{Sandbox: config.SandboxConfig{
		Runtime: "containerd",
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateSandbox_InvalidBackend(t *testing.T) {
	errs := config.Validate(config.Config{Sandbox: config.SandboxConfig{
		Backend: "k8s",
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateSandbox_InvalidMode(t *testing.T) {
	errs := config.Validate(config.Config{Sandbox: config.SandboxConfig{
		Mode: "shared",
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateGateway_ValidValues(t *testing.T) {
	errs := config.Validate(config.Config{Gateway: config.GatewayConfig{
		Port:     8080,
		Provider: "vertex",
	}})
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateGateway_InvalidProvider(t *testing.T) {
	errs := config.Validate(config.Config{Gateway: config.GatewayConfig{
		Provider: "azure",
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateGateway_InvalidPort(t *testing.T) {
	errs := config.Validate(config.Config{Gateway: config.GatewayConfig{
		Port: -1,
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateGateway_PortTooHigh(t *testing.T) {
	errs := config.Validate(config.Config{Gateway: config.GatewayConfig{
		Port: 70000,
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateEmbedding_ValidValues(t *testing.T) {
	errs := config.Validate(config.Config{Embedding: config.EmbeddingConfig{
		Provider:   "ollama",
		Dimensions: 256,
	}})
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateEmbedding_InvalidProvider(t *testing.T) {
	errs := config.Validate(config.Config{Embedding: config.EmbeddingConfig{
		Provider: "openai",
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateEmbedding_NegativeDimensions(t *testing.T) {
	errs := config.Validate(config.Config{Embedding: config.EmbeddingConfig{
		Dimensions: -1,
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateDoctor_ValidValues(t *testing.T) {
	errs := config.Validate(config.Config{Doctor: config.DoctorConfig{
		Tools: map[string]string{
			"gaze":    "recommended",
			"ollama":  "optional",
			"dewey":   "required",
		},
	}})
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateDoctor_InvalidSeverity(t *testing.T) {
	errs := config.Validate(config.Config{Doctor: config.DoctorConfig{
		Tools: map[string]string{
			"gaze": "critical",
		},
	}})
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidateDoctor_EmptyTools(t *testing.T) {
	errs := config.Validate(config.Config{Doctor: config.DoctorConfig{}})
	if len(errs) > 0 {
		t.Errorf("expected no errors for empty tools, got: %v", errs)
	}
}

// --- newConfigCmd tests ---

func TestNewConfigCmd_HasSubcommands(t *testing.T) {
	cmd := newConfigCmd()

	if cmd.Use != "config" {
		t.Errorf("Use = %q, want %q", cmd.Use, "config")
	}

	subcommands := cmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subcommands {
		names[sub.Use] = true
	}

	for _, want := range []string{"init", "show", "validate"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestConfigInitCmd_Execute(t *testing.T) {
	dir := t.TempDir()
	cmd := newConfigCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config init error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created") {
		t.Errorf("expected 'Created' in output, got: %s", output)
	}
}

func TestConfigShowCmd_Execute(t *testing.T) {
	dir := t.TempDir()
	cmd := newConfigCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"show", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config show error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "package_manager") {
		t.Errorf("expected 'package_manager' in output, got: %s", output)
	}
}

func TestConfigValidateCmd_Execute(t *testing.T) {
	dir := t.TempDir()
	cmd := newConfigCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"validate", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config validate error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No config file found") {
		t.Errorf("expected 'No config file found', got: %s", output)
	}
}

// --- Validate with zero values (empty structs) ---

func TestValidateSetup_EmptyIsValid(t *testing.T) {
	errs := config.Validate(config.Config{Setup: config.SetupConfig{}})
	if len(errs) > 0 {
		t.Errorf("expected no errors for empty setup, got: %v", errs)
	}
}

func TestValidateSandbox_EmptyIsValid(t *testing.T) {
	errs := config.Validate(config.Config{Sandbox: config.SandboxConfig{}})
	if len(errs) > 0 {
		t.Errorf("expected no errors for empty sandbox, got: %v", errs)
	}
}

func TestValidateGateway_EmptyIsValid(t *testing.T) {
	errs := config.Validate(config.Config{Gateway: config.GatewayConfig{}})
	if len(errs) > 0 {
		t.Errorf("expected no errors for empty gateway, got: %v", errs)
	}
}

func TestValidateEmbedding_EmptyIsValid(t *testing.T) {
	errs := config.Validate(config.Config{Embedding: config.EmbeddingConfig{}})
	if len(errs) > 0 {
		t.Errorf("expected no errors for empty embedding, got: %v", errs)
	}
}
