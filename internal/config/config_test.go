// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- Defaults tests ---

func TestDefaults_ReturnsPopulatedConfig(t *testing.T) {
	cfg := Defaults()

	if cfg.Setup.PackageManager != "auto" {
		t.Errorf("Setup.PackageManager = %q, want %q", cfg.Setup.PackageManager, "auto")
	}
	if cfg.Scaffold.Language != "auto" {
		t.Errorf("Scaffold.Language = %q, want %q", cfg.Scaffold.Language, "auto")
	}
	if cfg.Embedding.Model != "granite-embedding:30m" {
		t.Errorf("Embedding.Model = %q, want %q", cfg.Embedding.Model, "granite-embedding:30m")
	}
	if cfg.Embedding.Dimensions != 256 {
		t.Errorf("Embedding.Dimensions = %d, want %d", cfg.Embedding.Dimensions, 256)
	}
	if cfg.Embedding.Provider != "ollama" {
		t.Errorf("Embedding.Provider = %q, want %q", cfg.Embedding.Provider, "ollama")
	}
	if cfg.Embedding.Host != "http://localhost:11434" {
		t.Errorf("Embedding.Host = %q, want %q", cfg.Embedding.Host, "http://localhost:11434")
	}
	if cfg.Sandbox.Image != "quay.io/unbound-force/opencode-dev:latest" {
		t.Errorf("Sandbox.Image = %q, want %q", cfg.Sandbox.Image, "quay.io/unbound-force/opencode-dev:latest")
	}
	if cfg.Sandbox.Resources.Memory != "8g" {
		t.Errorf("Sandbox.Resources.Memory = %q, want %q", cfg.Sandbox.Resources.Memory, "8g")
	}
	if cfg.Sandbox.Resources.CPUs != "4" {
		t.Errorf("Sandbox.Resources.CPUs = %q, want %q", cfg.Sandbox.Resources.CPUs, "4")
	}
	if cfg.Sandbox.Mode != "isolated" {
		t.Errorf("Sandbox.Mode = %q, want %q", cfg.Sandbox.Mode, "isolated")
	}
	if cfg.Gateway.Port != 53147 {
		t.Errorf("Gateway.Port = %d, want %d", cfg.Gateway.Port, 53147)
	}
	if cfg.Gateway.Provider != "auto" {
		t.Errorf("Gateway.Provider = %q, want %q", cfg.Gateway.Provider, "auto")
	}
	if cfg.Workflow.ExecutionModes["define"] != "human" {
		t.Errorf("Workflow.ExecutionModes[define] = %q, want %q", cfg.Workflow.ExecutionModes["define"], "human")
	}
	if cfg.Workflow.ExecutionModes["implement"] != "swarm" {
		t.Errorf("Workflow.ExecutionModes[implement] = %q, want %q", cfg.Workflow.ExecutionModes["implement"], "swarm")
	}
}

// --- Merge tests ---

func TestMerge_NonZeroOverlayWins(t *testing.T) {
	base := Defaults()
	overlay := Config{
		Setup: SetupConfig{PackageManager: "dnf"},
	}
	result := merge(base, overlay)
	if result.Setup.PackageManager != "dnf" {
		t.Errorf("PackageManager = %q, want %q", result.Setup.PackageManager, "dnf")
	}
	// Other fields preserved from base.
	if result.Embedding.Model != "granite-embedding:30m" {
		t.Errorf("Embedding.Model = %q, want %q", result.Embedding.Model, "granite-embedding:30m")
	}
}

func TestMerge_ZeroOverlayPreservesBase(t *testing.T) {
	base := Defaults()
	overlay := Config{} // all zero
	result := merge(base, overlay)
	if result.Setup.PackageManager != "auto" {
		t.Errorf("PackageManager = %q, want %q", result.Setup.PackageManager, "auto")
	}
	if result.Gateway.Port != 53147 {
		t.Errorf("Gateway.Port = %d, want %d", result.Gateway.Port, 53147)
	}
}

func TestMerge_SliceReplacement(t *testing.T) {
	base := Config{
		Setup: SetupConfig{Skip: []string{"a", "b"}},
	}
	overlay := Config{
		Setup: SetupConfig{Skip: []string{"c"}},
	}
	result := merge(base, overlay)
	if len(result.Setup.Skip) != 1 || result.Setup.Skip[0] != "c" {
		t.Errorf("Skip = %v, want [c]", result.Setup.Skip)
	}
}

func TestMerge_MapMerging(t *testing.T) {
	base := Config{
		Setup: SetupConfig{
			Tools: map[string]ToolConfig{
				"gaze":     {Method: "homebrew"},
				"opencode": {Method: "curl"},
			},
		},
	}
	overlay := Config{
		Setup: SetupConfig{
			Tools: map[string]ToolConfig{
				"gaze": {Method: "rpm"},
				"node": {Method: "fnm", Version: "22"},
			},
		},
	}
	result := merge(base, overlay)
	if result.Setup.Tools["gaze"].Method != "rpm" {
		t.Errorf("Tools[gaze].Method = %q, want %q", result.Setup.Tools["gaze"].Method, "rpm")
	}
	if result.Setup.Tools["opencode"].Method != "curl" {
		t.Errorf("Tools[opencode].Method = %q, want %q", result.Setup.Tools["opencode"].Method, "curl")
	}
	if result.Setup.Tools["node"].Method != "fnm" {
		t.Errorf("Tools[node].Method = %q, want %q", result.Setup.Tools["node"].Method, "fnm")
	}
	if result.Setup.Tools["node"].Version != "22" {
		t.Errorf("Tools[node].Version = %q, want %q", result.Setup.Tools["node"].Version, "22")
	}
}

func TestMerge_DeepFields(t *testing.T) {
	base := Defaults()
	overlay := Config{
		Sandbox: SandboxConfig{
			Resources: ResourcesConfig{Memory: "16g"},
		},
	}
	result := merge(base, overlay)
	if result.Sandbox.Resources.Memory != "16g" {
		t.Errorf("Resources.Memory = %q, want %q", result.Sandbox.Resources.Memory, "16g")
	}
	if result.Sandbox.Resources.CPUs != "4" {
		t.Errorf("Resources.CPUs = %q, want %q (preserved from base)", result.Sandbox.Resources.CPUs, "4")
	}
}

func TestMerge_WorkflowModes(t *testing.T) {
	base := Defaults()
	overlay := Config{
		Workflow: WorkflowConfig{
			ExecutionModes: map[string]string{
				"define": "swarm",
			},
			SpecReview: true,
		},
	}
	result := merge(base, overlay)
	if result.Workflow.ExecutionModes["define"] != "swarm" {
		t.Errorf("ExecutionModes[define] = %q, want %q", result.Workflow.ExecutionModes["define"], "swarm")
	}
	if result.Workflow.ExecutionModes["implement"] != "swarm" {
		t.Errorf("ExecutionModes[implement] = %q, want %q (preserved from base)", result.Workflow.ExecutionModes["implement"], "swarm")
	}
	if !result.Workflow.SpecReview {
		t.Error("SpecReview = false, want true")
	}
}

// --- EnvOverrides tests ---

func TestApplyEnvOverrides_StringFields(t *testing.T) {
	cfg := Defaults()
	env := map[string]string{
		"UF_PACKAGE_MANAGER":  "dnf",
		"OLLAMA_MODEL":        "mxbai-embed-large",
		"OLLAMA_HOST":         "http://remote:11434",
		"UF_SANDBOX_IMAGE":    "custom:v2",
		"UF_SANDBOX_BACKEND":  "podman",
		"UF_SANDBOX_RUNTIME":  "docker",
		"UF_GATEWAY_PROVIDER": "vertex",
	}
	result := applyEnvOverrides(cfg, func(k string) string { return env[k] })

	checks := []struct {
		name, got, want string
	}{
		{"PackageManager", result.Setup.PackageManager, "dnf"},
		{"Embedding.Model", result.Embedding.Model, "mxbai-embed-large"},
		{"Embedding.Host", result.Embedding.Host, "http://remote:11434"},
		{"Sandbox.Image", result.Sandbox.Image, "custom:v2"},
		{"Sandbox.Backend", result.Sandbox.Backend, "podman"},
		{"Sandbox.Runtime", result.Sandbox.Runtime, "docker"},
		{"Gateway.Provider", result.Gateway.Provider, "vertex"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestApplyEnvOverrides_IntFields(t *testing.T) {
	cfg := Defaults()
	env := map[string]string{
		"OLLAMA_EMBED_DIM":  "1024",
		"UF_GATEWAY_PORT":   "9999",
	}
	result := applyEnvOverrides(cfg, func(k string) string { return env[k] })

	if result.Embedding.Dimensions != 1024 {
		t.Errorf("Embedding.Dimensions = %d, want %d", result.Embedding.Dimensions, 1024)
	}
	if result.Gateway.Port != 9999 {
		t.Errorf("Gateway.Port = %d, want %d", result.Gateway.Port, 9999)
	}
}

func TestApplyEnvOverrides_InvalidIntIgnored(t *testing.T) {
	cfg := Defaults()
	env := map[string]string{
		"UF_GATEWAY_PORT": "notanumber",
	}
	result := applyEnvOverrides(cfg, func(k string) string { return env[k] })
	if result.Gateway.Port != 53147 {
		t.Errorf("Gateway.Port = %d, want %d (unchanged)", result.Gateway.Port, 53147)
	}
}

func TestApplyEnvOverrides_EmptyEnvPreservesDefaults(t *testing.T) {
	cfg := Defaults()
	result := applyEnvOverrides(cfg, func(string) string { return "" })
	if result.Embedding.Model != "granite-embedding:30m" {
		t.Errorf("Embedding.Model = %q, want %q", result.Embedding.Model, "granite-embedding:30m")
	}
}

// --- Load tests ---

func TestLoad_NoFiles(t *testing.T) {
	cfg, err := Load(LoadOptions{
		ProjectDir: t.TempDir(),
		ReadFile:   func(string) ([]byte, error) { return nil, os.ErrNotExist },
		Getenv:     func(string) string { return "" },
		UserConfigDir: func() (string, error) {
			return t.TempDir(), nil
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	defaults := Defaults()
	if cfg.Embedding.Model != defaults.Embedding.Model {
		t.Errorf("Embedding.Model = %q, want %q", cfg.Embedding.Model, defaults.Embedding.Model)
	}
	if cfg.Gateway.Port != defaults.Gateway.Port {
		t.Errorf("Gateway.Port = %d, want %d", cfg.Gateway.Port, defaults.Gateway.Port)
	}
}

func TestLoad_RepoConfigOnly(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte("sandbox:\n  runtime: podman\n  image: custom:v3\n")
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(LoadOptions{
		ProjectDir: dir,
		Getenv:     func(string) string { return "" },
		UserConfigDir: func() (string, error) {
			return filepath.Join(dir, "no-user-config"), nil
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Sandbox.Runtime != "podman" {
		t.Errorf("Sandbox.Runtime = %q, want %q", cfg.Sandbox.Runtime, "podman")
	}
	if cfg.Sandbox.Image != "custom:v3" {
		t.Errorf("Sandbox.Image = %q, want %q", cfg.Sandbox.Image, "custom:v3")
	}
	// Other defaults preserved.
	if cfg.Embedding.Model != "granite-embedding:30m" {
		t.Errorf("Embedding.Model = %q, want %q", cfg.Embedding.Model, "granite-embedding:30m")
	}
}

func TestLoad_UserConfigOnly(t *testing.T) {
	projectDir := t.TempDir()
	userDir := t.TempDir()
	ufUserDir := filepath.Join(userDir, "uf")
	if err := os.MkdirAll(ufUserDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte("setup:\n  package_manager: dnf\n")
	if err := os.WriteFile(filepath.Join(ufUserDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(LoadOptions{
		ProjectDir:    projectDir,
		Getenv:        func(string) string { return "" },
		UserConfigDir: func() (string, error) { return userDir, nil },
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Setup.PackageManager != "dnf" {
		t.Errorf("Setup.PackageManager = %q, want %q", cfg.Setup.PackageManager, "dnf")
	}
}

func TestLoad_RepoOverridesUser(t *testing.T) {
	projectDir := t.TempDir()
	userDir := t.TempDir()

	// User config: sandbox.runtime = docker
	ufUserDir := filepath.Join(userDir, "uf")
	if err := os.MkdirAll(ufUserDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(ufUserDir, "config.yaml"),
		[]byte("sandbox:\n  runtime: docker\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// Repo config: sandbox.runtime = podman
	ufRepoDir := filepath.Join(projectDir, ".uf")
	if err := os.MkdirAll(ufRepoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(ufRepoDir, "config.yaml"),
		[]byte("sandbox:\n  runtime: podman\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(LoadOptions{
		ProjectDir:    projectDir,
		Getenv:        func(string) string { return "" },
		UserConfigDir: func() (string, error) { return userDir, nil },
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Sandbox.Runtime != "podman" {
		t.Errorf("Sandbox.Runtime = %q, want %q (repo should override user)", cfg.Sandbox.Runtime, "podman")
	}
}

func TestLoad_EnvOverridesConfig(t *testing.T) {
	projectDir := t.TempDir()
	ufDir := filepath.Join(projectDir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(ufDir, "config.yaml"),
		[]byte("sandbox:\n  image: from-config:v1\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(LoadOptions{
		ProjectDir: projectDir,
		Getenv: func(k string) string {
			if k == "UF_SANDBOX_IMAGE" {
				return "from-env:v2"
			}
			return ""
		},
		UserConfigDir: func() (string, error) {
			return filepath.Join(projectDir, "no-user"), nil
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Sandbox.Image != "from-env:v2" {
		t.Errorf("Sandbox.Image = %q, want %q (env should override config)", cfg.Sandbox.Image, "from-env:v2")
	}
}

func TestLoad_MergePreservesNonOverlapping(t *testing.T) {
	projectDir := t.TempDir()
	userDir := t.TempDir()

	// User sets package_manager.
	ufUserDir := filepath.Join(userDir, "uf")
	if err := os.MkdirAll(ufUserDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(ufUserDir, "config.yaml"),
		[]byte("setup:\n  package_manager: dnf\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// Repo sets sandbox.runtime.
	ufRepoDir := filepath.Join(projectDir, ".uf")
	if err := os.MkdirAll(ufRepoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(ufRepoDir, "config.yaml"),
		[]byte("sandbox:\n  runtime: podman\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(LoadOptions{
		ProjectDir:    projectDir,
		Getenv:        func(string) string { return "" },
		UserConfigDir: func() (string, error) { return userDir, nil },
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Setup.PackageManager != "dnf" {
		t.Errorf("PackageManager = %q, want %q", cfg.Setup.PackageManager, "dnf")
	}
	if cfg.Sandbox.Runtime != "podman" {
		t.Errorf("Sandbox.Runtime = %q, want %q", cfg.Sandbox.Runtime, "podman")
	}
}

// --- UIDMap tests (sandbox-uid-mapping change, 8.20, 8.21) ---

func TestApplyEnvOverrides_UIDMap(t *testing.T) {
	// UF_SANDBOX_UIDMAP=1 sets UIDMap to true.
	cfg := Defaults()
	result := applyEnvOverrides(cfg, func(k string) string {
		if k == "UF_SANDBOX_UIDMAP" {
			return "1"
		}
		return ""
	})
	if !result.Sandbox.UIDMap {
		t.Error("expected Sandbox.UIDMap=true when UF_SANDBOX_UIDMAP=1")
	}

	// UF_SANDBOX_UIDMAP=true also works.
	cfg2 := Defaults()
	result2 := applyEnvOverrides(cfg2, func(k string) string {
		if k == "UF_SANDBOX_UIDMAP" {
			return "true"
		}
		return ""
	})
	if !result2.Sandbox.UIDMap {
		t.Error("expected Sandbox.UIDMap=true when UF_SANDBOX_UIDMAP=true")
	}

	// UF_SANDBOX_UIDMAP=0 leaves it false.
	cfg3 := Defaults()
	result3 := applyEnvOverrides(cfg3, func(k string) string {
		if k == "UF_SANDBOX_UIDMAP" {
			return "0"
		}
		return ""
	})
	if result3.Sandbox.UIDMap {
		t.Error("expected Sandbox.UIDMap=false when UF_SANDBOX_UIDMAP=0")
	}

	// No env var leaves it false.
	cfg4 := Defaults()
	result4 := applyEnvOverrides(cfg4, func(string) string { return "" })
	if result4.Sandbox.UIDMap {
		t.Error("expected Sandbox.UIDMap=false when env var not set")
	}
}

func TestConfigMerge_UIDMap(t *testing.T) {
	base := Defaults()
	overlay := Config{
		Sandbox: SandboxConfig{UIDMap: true},
	}
	result := merge(base, overlay)
	if !result.Sandbox.UIDMap {
		t.Error("expected UIDMap=true after merge with overlay UIDMap=true")
	}

	// Zero overlay preserves base.
	base2 := Defaults()
	base2.Sandbox.UIDMap = false
	overlay2 := Config{}
	result2 := merge(base2, overlay2)
	if result2.Sandbox.UIDMap {
		t.Error("expected UIDMap=false when overlay has zero value")
	}
}

// --- SandboxConfig.IsEmpty tests ---

func TestSandboxConfig_IsEmpty(t *testing.T) {
	if !Defaults().Sandbox.IsEmpty() {
		t.Error("default sandbox config should be considered empty")
	}

	custom := Defaults().Sandbox
	custom.DemoPorts = []int{3000}
	if custom.IsEmpty() {
		t.Error("sandbox config with DemoPorts should not be empty")
	}
}

// --- Path helpers ---

func TestRepoConfigPath(t *testing.T) {
	got := RepoConfigPath("/project")
	want := filepath.Join("/project", ".uf", "config.yaml")
	if got != want {
		t.Errorf("RepoConfigPath = %q, want %q", got, want)
	}
}

func TestUserConfigPath(t *testing.T) {
	got, err := UserConfigPath(func() (string, error) {
		return "/home/user/.config", nil
	})
	if err != nil {
		t.Fatalf("UserConfigPath error = %v", err)
	}
	want := filepath.Join("/home/user/.config", "uf", "config.yaml")
	if got != want {
		t.Errorf("UserConfigPath = %q, want %q", got, want)
	}
}

func TestUserConfigPath_Error(t *testing.T) {
	_, err := UserConfigPath(func() (string, error) {
		return "", fmt.Errorf("no home")
	})
	if err == nil {
		t.Error("expected error when UserConfigDir fails")
	}
}

// --- JSON serialization test ---

func TestConfig_JSONRoundTrip(t *testing.T) {
	cfg := Defaults()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}
	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if decoded.Embedding.Model != cfg.Embedding.Model {
		t.Errorf("JSON round-trip: Embedding.Model = %q, want %q", decoded.Embedding.Model, cfg.Embedding.Model)
	}
	if decoded.Gateway.Port != cfg.Gateway.Port {
		t.Errorf("JSON round-trip: Gateway.Port = %d, want %d", decoded.Gateway.Port, cfg.Gateway.Port)
	}
}
