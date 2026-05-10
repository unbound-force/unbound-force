// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"testing"
)

// --- Validate (top-level orchestrator) tests ---

func TestValidate_ValidDefaults(t *testing.T) {
	cfg := Defaults()
	errs := Validate(cfg)
	if len(errs) != 0 {
		t.Errorf("Validate(Defaults()) = %v, want no errors", errs)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := Config{
		Setup:   SetupConfig{PackageManager: "invalid-pm"},
		Sandbox: SandboxConfig{Runtime: "invalid-rt"},
		Gateway: GatewayConfig{Provider: "invalid-pv", Port: -1},
	}
	errs := Validate(cfg)
	if len(errs) < 3 {
		t.Errorf("Validate() returned %d errors, want at least 3: %v", len(errs), errs)
	}
}

// --- validateSetup tests ---

func TestValidateSetup_ValidPackageManagers(t *testing.T) {
	valid := []string{"auto", "homebrew", "dnf", "apt", "manual", ""}
	for _, pm := range valid {
		t.Run("pm="+pm, func(t *testing.T) {
			cfg := SetupConfig{PackageManager: pm}
			errs := validateSetup(cfg)
			if len(errs) != 0 {
				t.Errorf("validateSetup(%q) = %v, want no errors", pm, errs)
			}
		})
	}
}

func TestValidateSetup_InvalidPackageManager(t *testing.T) {
	cfg := SetupConfig{PackageManager: "yum"}
	errs := validateSetup(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSetup(yum) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "setup.package_manager") {
		t.Errorf("error = %q, want to contain 'setup.package_manager'", errs[0])
	}
	if !strings.Contains(errs[0], "yum") {
		t.Errorf("error = %q, want to contain 'yum'", errs[0])
	}
}

func TestValidateSetup_ValidToolMethods(t *testing.T) {
	valid := []string{"auto", "homebrew", "dnf", "rpm", "apt", "curl", "skip", "nvm", "fnm", "mise", ""}
	for _, method := range valid {
		t.Run("method="+method, func(t *testing.T) {
			cfg := SetupConfig{
				Tools: map[string]ToolConfig{
					"test-tool": {Method: method},
				},
			}
			errs := validateSetup(cfg)
			if len(errs) != 0 {
				t.Errorf("validateSetup(method=%q) = %v, want no errors", method, errs)
			}
		})
	}
}

func TestValidateSetup_InvalidToolMethod(t *testing.T) {
	cfg := SetupConfig{
		Tools: map[string]ToolConfig{
			"my-tool": {Method: "invalid-method"},
		},
	}
	errs := validateSetup(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSetup() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "setup.tools.my-tool.method") {
		t.Errorf("error = %q, want to contain 'setup.tools.my-tool.method'", errs[0])
	}
	if !strings.Contains(errs[0], "invalid-method") {
		t.Errorf("error = %q, want to contain 'invalid-method'", errs[0])
	}
}

func TestValidateSetup_MultipleInvalidTools(t *testing.T) {
	cfg := SetupConfig{
		Tools: map[string]ToolConfig{
			"tool-a": {Method: "bad1"},
			"tool-b": {Method: "bad2"},
		},
	}
	errs := validateSetup(cfg)
	if len(errs) != 2 {
		t.Errorf("validateSetup() returned %d errors, want 2: %v", len(errs), errs)
	}
}

// --- validateSandbox tests ---

func TestValidateSandbox_AllValidCombinations(t *testing.T) {
	tests := []struct {
		name    string
		runtime string
		backend string
		mode    string
	}{
		{"all empty", "", "", ""},
		{"auto runtime", "auto", "auto", "isolated"},
		{"podman/podman/direct", "podman", "podman", "direct"},
		{"docker/che/isolated", "docker", "che", "isolated"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := SandboxConfig{
				Runtime: tc.runtime,
				Backend: tc.backend,
				Mode:    tc.mode,
			}
			errs := validateSandbox(cfg)
			if len(errs) != 0 {
				t.Errorf("validateSandbox() = %v, want no errors", errs)
			}
		})
	}
}

func TestValidateSandbox_InvalidRuntime(t *testing.T) {
	cfg := SandboxConfig{Runtime: "containerd"}
	errs := validateSandbox(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSandbox() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "sandbox.runtime") {
		t.Errorf("error = %q, want to contain 'sandbox.runtime'", errs[0])
	}
	if !strings.Contains(errs[0], "containerd") {
		t.Errorf("error = %q, want to contain 'containerd'", errs[0])
	}
}

func TestValidateSandbox_InvalidBackend(t *testing.T) {
	cfg := SandboxConfig{Backend: "kubernetes"}
	errs := validateSandbox(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSandbox() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "sandbox.backend") {
		t.Errorf("error = %q, want to contain 'sandbox.backend'", errs[0])
	}
}

func TestValidateSandbox_InvalidMode(t *testing.T) {
	cfg := SandboxConfig{Mode: "hybrid"}
	errs := validateSandbox(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSandbox() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "sandbox.mode") {
		t.Errorf("error = %q, want to contain 'sandbox.mode'", errs[0])
	}
}

func TestValidateSandbox_MultipleErrors(t *testing.T) {
	cfg := SandboxConfig{
		Runtime: "bad-rt",
		Backend: "bad-be",
		Mode:    "bad-mode",
	}
	errs := validateSandbox(cfg)
	if len(errs) != 3 {
		t.Errorf("validateSandbox() returned %d errors, want 3: %v", len(errs), errs)
	}
}

// --- validateGateway tests ---

func TestValidateGateway_ValidProviders(t *testing.T) {
	valid := []string{"auto", "anthropic", "vertex", "bedrock", ""}
	for _, pv := range valid {
		t.Run("provider="+pv, func(t *testing.T) {
			cfg := GatewayConfig{Provider: pv, Port: 8080}
			errs := validateGateway(cfg)
			if len(errs) != 0 {
				t.Errorf("validateGateway(%q) = %v, want no errors", pv, errs)
			}
		})
	}
}

func TestValidateGateway_InvalidProvider(t *testing.T) {
	cfg := GatewayConfig{Provider: "openai"}
	errs := validateGateway(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateGateway() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "gateway.provider") {
		t.Errorf("error = %q, want to contain 'gateway.provider'", errs[0])
	}
	if !strings.Contains(errs[0], "openai") {
		t.Errorf("error = %q, want to contain 'openai'", errs[0])
	}
}

func TestValidateGateway_ValidPortRange(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero (valid)", 0},
		{"standard", 8080},
		{"max", 65535},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := GatewayConfig{Port: tc.port}
			errs := validateGateway(cfg)
			if len(errs) != 0 {
				t.Errorf("validateGateway(port=%d) = %v, want no errors", tc.port, errs)
			}
		})
	}
}

func TestValidateGateway_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"negative", -1},
		{"too large", 65536},
		{"very negative", -100},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := GatewayConfig{Port: tc.port}
			errs := validateGateway(cfg)
			if len(errs) != 1 {
				t.Fatalf("validateGateway(port=%d) returned %d errors, want 1", tc.port, len(errs))
			}
			if !strings.Contains(errs[0], "gateway.port") {
				t.Errorf("error = %q, want to contain 'gateway.port'", errs[0])
			}
		})
	}
}

func TestValidateGateway_InvalidProviderAndPort(t *testing.T) {
	cfg := GatewayConfig{Provider: "bad", Port: -1}
	errs := validateGateway(cfg)
	if len(errs) != 2 {
		t.Errorf("validateGateway() returned %d errors, want 2: %v", len(errs), errs)
	}
}

// --- validateEmbedding tests ---

func TestValidateEmbedding_ValidConfig(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		dims     int
	}{
		{"ollama", "ollama", 256},
		{"empty provider", "", 0},
		{"zero dimensions", "ollama", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := EmbeddingConfig{Provider: tc.provider, Dimensions: tc.dims}
			errs := validateEmbedding(cfg)
			if len(errs) != 0 {
				t.Errorf("validateEmbedding() = %v, want no errors", errs)
			}
		})
	}
}

func TestValidateEmbedding_InvalidProvider(t *testing.T) {
	cfg := EmbeddingConfig{Provider: "openai"}
	errs := validateEmbedding(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateEmbedding() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "embedding.provider") {
		t.Errorf("error = %q, want to contain 'embedding.provider'", errs[0])
	}
}

func TestValidateEmbedding_NegativeDimensions(t *testing.T) {
	cfg := EmbeddingConfig{Provider: "ollama", Dimensions: -1}
	errs := validateEmbedding(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateEmbedding() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "embedding.dimensions") {
		t.Errorf("error = %q, want to contain 'embedding.dimensions'", errs[0])
	}
}

// --- validateDoctor tests ---

func TestValidateDoctor_ValidSeverities(t *testing.T) {
	cfg := DoctorConfig{
		Tools: map[string]string{
			"gaze":     "required",
			"opencode": "recommended",
			"gh":       "optional",
		},
	}
	errs := validateDoctor(cfg)
	if len(errs) != 0 {
		t.Errorf("validateDoctor() = %v, want no errors", errs)
	}
}

func TestValidateDoctor_InvalidSeverity(t *testing.T) {
	cfg := DoctorConfig{
		Tools: map[string]string{
			"gaze": "critical",
		},
	}
	errs := validateDoctor(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateDoctor() returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "doctor.tools.gaze") {
		t.Errorf("error = %q, want to contain 'doctor.tools.gaze'", errs[0])
	}
	if !strings.Contains(errs[0], "critical") {
		t.Errorf("error = %q, want to contain 'critical'", errs[0])
	}
}

func TestValidateDoctor_EmptyTools(t *testing.T) {
	cfg := DoctorConfig{Tools: nil}
	errs := validateDoctor(cfg)
	if len(errs) != 0 {
		t.Errorf("validateDoctor(nil tools) = %v, want no errors", errs)
	}
}
