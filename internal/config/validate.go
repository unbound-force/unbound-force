// SPDX-License-Identifier: Apache-2.0

package config

import "fmt"

// Validate checks a Config for invalid field values. Returns
// a slice of human-readable error strings, one per invalid field.
// An empty slice means the config is valid.
func Validate(cfg Config) []string {
	var errs []string
	errs = append(errs, validateSetup(cfg.Setup)...)
	errs = append(errs, validateSandbox(cfg.Sandbox)...)
	errs = append(errs, validateGateway(cfg.Gateway)...)
	errs = append(errs, validateEmbedding(cfg.Embedding)...)
	errs = append(errs, validateDoctor(cfg.Doctor)...)
	return errs
}

func validateSetup(cfg SetupConfig) []string {
	var errs []string
	valid := map[string]bool{
		"auto": true, "homebrew": true, "dnf": true,
		"apt": true, "manual": true, "": true,
	}
	if !valid[cfg.PackageManager] {
		errs = append(errs, fmt.Sprintf(
			"setup.package_manager: %q is not valid (auto|homebrew|dnf|apt|manual)",
			cfg.PackageManager))
	}

	validMethods := map[string]bool{
		"auto": true, "homebrew": true, "dnf": true, "rpm": true,
		"apt": true, "curl": true, "skip": true, "nvm": true,
		"fnm": true, "mise": true, "": true,
	}
	for name, tool := range cfg.Tools {
		if !validMethods[tool.Method] {
			errs = append(errs, fmt.Sprintf(
				"setup.tools.%s.method: %q is not valid",
				name, tool.Method))
		}
	}
	return errs
}

func validateSandbox(cfg SandboxConfig) []string {
	var errs []string
	validRuntime := map[string]bool{
		"auto": true, "podman": true, "docker": true, "": true,
	}
	if !validRuntime[cfg.Runtime] {
		errs = append(errs, fmt.Sprintf(
			"sandbox.runtime: %q is not valid (auto|podman|docker)",
			cfg.Runtime))
	}

	validBackend := map[string]bool{
		"auto": true, "podman": true, "che": true, "": true,
	}
	if !validBackend[cfg.Backend] {
		errs = append(errs, fmt.Sprintf(
			"sandbox.backend: %q is not valid (auto|podman|che)",
			cfg.Backend))
	}

	validMode := map[string]bool{
		"isolated": true, "direct": true, "": true,
	}
	if !validMode[cfg.Mode] {
		errs = append(errs, fmt.Sprintf(
			"sandbox.mode: %q is not valid (isolated|direct)",
			cfg.Mode))
	}
	return errs
}

func validateGateway(cfg GatewayConfig) []string {
	var errs []string
	validProvider := map[string]bool{
		"auto": true, "anthropic": true, "vertex": true,
		"bedrock": true, "": true,
	}
	if !validProvider[cfg.Provider] {
		errs = append(errs, fmt.Sprintf(
			"gateway.provider: %q is not valid (auto|anthropic|vertex|bedrock)",
			cfg.Provider))
	}
	if cfg.Port < 0 || cfg.Port > 65535 {
		errs = append(errs, fmt.Sprintf(
			"gateway.port: %d is not a valid port number (0-65535)",
			cfg.Port))
	}
	return errs
}

func validateEmbedding(cfg EmbeddingConfig) []string {
	var errs []string
	validProvider := map[string]bool{
		"ollama": true, "": true,
	}
	if !validProvider[cfg.Provider] {
		errs = append(errs, fmt.Sprintf(
			"embedding.provider: %q is not valid (ollama)",
			cfg.Provider))
	}
	if cfg.Dimensions < 0 {
		errs = append(errs, fmt.Sprintf(
			"embedding.dimensions: %d must be non-negative",
			cfg.Dimensions))
	}
	return errs
}

func validateDoctor(cfg DoctorConfig) []string {
	var errs []string
	validSeverity := map[string]bool{
		"required": true, "recommended": true, "optional": true,
	}
	for name, sev := range cfg.Tools {
		if !validSeverity[sev] {
			errs = append(errs, fmt.Sprintf(
				"doctor.tools.%s: %q is not valid (required|recommended|optional)",
				name, sev))
		}
	}
	return errs
}
