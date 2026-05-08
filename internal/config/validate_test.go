// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"testing"
)

// --- validateSetup tests ---

func TestValidateSetup_ValidPackageManagers(t *testing.T) {
	validManagers := []string{"auto", "homebrew", "dnf", "apt", "manual", ""}
	for _, pm := range validManagers {
		cfg := SetupConfig{PackageManager: pm}
		errs := validateSetup(cfg)
		if len(errs) != 0 {
			t.Errorf("validateSetup(%q) returned errors: %v", pm, errs)
		}
	}
}

func TestValidateSetup_InvalidPackageManager(t *testing.T) {
	cfg := SetupConfig{PackageManager: "yum"}
	errs := validateSetup(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSetup(yum) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "yum") {
		t.Errorf("error = %q, want to contain %q", errs[0], "yum")
	}
	if !strings.Contains(errs[0], "setup.package_manager") {
		t.Errorf("error = %q, want to contain field path", errs[0])
	}
}

func TestValidateSetup_ValidToolMethods(t *testing.T) {
	validMethods := []string{"auto", "homebrew", "dnf", "rpm", "apt", "curl", "skip", "nvm", "fnm", "mise", ""}
	for _, method := range validMethods {
		cfg := SetupConfig{
			Tools: map[string]ToolConfig{
				"gaze": {Method: method},
			},
		}
		errs := validateSetup(cfg)
		if len(errs) != 0 {
			t.Errorf("validateSetup(tool method %q) returned errors: %v", method, errs)
		}
	}
}

func TestValidateSetup_InvalidToolMethod(t *testing.T) {
	cfg := SetupConfig{
		Tools: map[string]ToolConfig{
			"gaze": {Method: "snap"},
		},
	}
	errs := validateSetup(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSetup(tool method snap) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "gaze") {
		t.Errorf("error = %q, want to contain tool name", errs[0])
	}
	if !strings.Contains(errs[0], "snap") {
		t.Errorf("error = %q, want to contain invalid method", errs[0])
	}
}

func TestValidateSetup_MultipleToolErrors(t *testing.T) {
	cfg := SetupConfig{
		PackageManager: "invalid_pm",
		Tools: map[string]ToolConfig{
			"gaze":    {Method: "bad1"},
			"opencode": {Method: "bad2"},
		},
	}
	errs := validateSetup(cfg)
	// 1 for package manager + 2 for tools = 3
	if len(errs) != 3 {
		t.Fatalf("validateSetup returned %d errors, want 3; errors: %v", len(errs), errs)
	}
}

// --- validateSandbox tests ---

func TestValidateSandbox_ValidValues(t *testing.T) {
	tc := []SandboxConfig{
		{},
		{Runtime: "auto", Backend: "auto", Mode: "isolated"},
		{Runtime: "podman", Backend: "podman", Mode: "direct"},
		{Runtime: "docker", Backend: "che"},
	}
	for _, cfg := range tc {
		errs := validateSandbox(cfg)
		if len(errs) != 0 {
			t.Errorf("validateSandbox(%+v) returned errors: %v", cfg, errs)
		}
	}
}

func TestValidateSandbox_InvalidRuntime(t *testing.T) {
	cfg := SandboxConfig{Runtime: "containerd"}
	errs := validateSandbox(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSandbox(containerd) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "sandbox.runtime") {
		t.Errorf("error = %q, want to contain field path", errs[0])
	}
	if !strings.Contains(errs[0], "containerd") {
		t.Errorf("error = %q, want to contain invalid value", errs[0])
	}
}

func TestValidateSandbox_InvalidBackend(t *testing.T) {
	cfg := SandboxConfig{Backend: "kubernetes"}
	errs := validateSandbox(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSandbox(kubernetes backend) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "sandbox.backend") {
		t.Errorf("error = %q, want to contain field path", errs[0])
	}
}

func TestValidateSandbox_InvalidMode(t *testing.T) {
	cfg := SandboxConfig{Mode: "hybrid"}
	errs := validateSandbox(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateSandbox(hybrid mode) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "sandbox.mode") {
		t.Errorf("error = %q, want to contain field path", errs[0])
	}
}

func TestValidateSandbox_MultipleErrors(t *testing.T) {
	cfg := SandboxConfig{
		Runtime: "bad_runtime",
		Backend: "bad_backend",
		Mode:    "bad_mode",
	}
	errs := validateSandbox(cfg)
	if len(errs) != 3 {
		t.Fatalf("validateSandbox returned %d errors, want 3; errors: %v", len(errs), errs)
	}
}

// --- validateGateway tests ---

func TestValidateGateway_ValidValues(t *testing.T) {
	tc := []GatewayConfig{
		{},
		{Provider: "auto", Port: 8080},
		{Provider: "anthropic", Port: 0},
		{Provider: "vertex", Port: 65535},
		{Provider: "bedrock", Port: 443},
	}
	for _, cfg := range tc {
		errs := validateGateway(cfg)
		if len(errs) != 0 {
			t.Errorf("validateGateway(%+v) returned errors: %v", cfg, errs)
		}
	}
}

func TestValidateGateway_InvalidProvider(t *testing.T) {
	cfg := GatewayConfig{Provider: "openai"}
	errs := validateGateway(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateGateway(openai) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "gateway.provider") {
		t.Errorf("error = %q, want to contain field path", errs[0])
	}
	if !strings.Contains(errs[0], "openai") {
		t.Errorf("error = %q, want to contain invalid value", errs[0])
	}
}

func TestValidateGateway_InvalidPort_Negative(t *testing.T) {
	cfg := GatewayConfig{Port: -1}
	errs := validateGateway(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateGateway(port -1) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "gateway.port") {
		t.Errorf("error = %q, want to contain field path", errs[0])
	}
}

func TestValidateGateway_InvalidPort_TooHigh(t *testing.T) {
	cfg := GatewayConfig{Port: 70000}
	errs := validateGateway(cfg)
	if len(errs) != 1 {
		t.Fatalf("validateGateway(port 70000) returned %d errors, want 1", len(errs))
	}
	if !strings.Contains(errs[0], "70000") {
		t.Errorf("error = %q, want to contain invalid port", errs[0])
	}
}

func TestValidateGateway_MultipleErrors(t *testing.T) {
	cfg := GatewayConfig{Provider: "invalid", Port: -5}
	errs := validateGateway(cfg)
	if len(errs) != 2 {
		t.Fatalf("validateGateway returned %d errors, want 2; errors: %v", len(errs), errs)
	}
}

// --- Validate (top-level) tests ---

func TestValidate_ValidConfig(t *testing.T) {
	cfg := Defaults()
	errs := Validate(cfg)
	if len(errs) != 0 {
		t.Errorf("Validate(Defaults()) returned errors: %v", errs)
	}
}

func TestValidate_AccumulatesAllErrors(t *testing.T) {
	cfg := Config{
		Setup:   SetupConfig{PackageManager: "invalid_pm"},
		Sandbox: SandboxConfig{Runtime: "invalid_rt"},
		Gateway: GatewayConfig{Provider: "invalid_pv", Port: -1},
	}
	errs := Validate(cfg)
	// 1 (setup.package_manager) + 1 (sandbox.runtime) +
	// 2 (gateway.provider + gateway.port) = 4
	if len(errs) != 4 {
		t.Fatalf("Validate returned %d errors, want 4; errors: %v", len(errs), errs)
	}
}
