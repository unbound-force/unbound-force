// SPDX-License-Identifier: Apache-2.0

// Package config provides unified configuration loading for the
// Unbound Force CLI. It implements layered resolution:
//
//	CLI flags > env vars > repo config > user config > compiled defaults
//
// The config file at .uf/config.yaml is optional — missing files
// produce compiled defaults with no error.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/log"
	goyaml "github.com/goccy/go-yaml"
)

// Config is the unified configuration for the Unbound Force CLI.
// All fields use zero-value semantics: absent fields in YAML
// produce zero values, which the merge function treats as "not set."
type Config struct {
	Setup     SetupConfig     `yaml:"setup"     json:"setup"`
	Scaffold  ScaffoldConfig  `yaml:"scaffold"  json:"scaffold"`
	Embedding EmbeddingConfig `yaml:"embedding" json:"embedding"`
	Sandbox   SandboxConfig   `yaml:"sandbox"   json:"sandbox"`
	Gateway   GatewayConfig   `yaml:"gateway"   json:"gateway"`
	Doctor    DoctorConfig    `yaml:"doctor"    json:"doctor"`
	Workflow  WorkflowConfig  `yaml:"workflow"  json:"workflow"`
}

// SetupConfig controls how `uf setup` installs tools.
type SetupConfig struct {
	PackageManager string                `yaml:"package_manager" json:"package_manager"`
	Skip           []string              `yaml:"skip"            json:"skip"`
	Tools          map[string]ToolConfig `yaml:"tools"           json:"tools"`
}

// ToolConfig defines per-tool install method overrides.
type ToolConfig struct {
	Method  string `yaml:"method"  json:"method"`
	Version string `yaml:"version" json:"version,omitempty"`
}

// ScaffoldConfig controls what `uf init` deploys.
type ScaffoldConfig struct {
	Language string `yaml:"language" json:"language"`
}

// EmbeddingConfig controls the embedding model used by Dewey.
type EmbeddingConfig struct {
	Model      string `yaml:"model"      json:"model"`
	Dimensions int    `yaml:"dimensions" json:"dimensions"`
	Provider   string `yaml:"provider"   json:"provider"`
	Host       string `yaml:"host"       json:"host"`
}

// SandboxConfig controls `uf sandbox` behavior. Absorbs the
// previously separate .uf/sandbox.yaml file.
type SandboxConfig struct {
	Runtime   string          `yaml:"runtime"    json:"runtime"`
	Backend   string          `yaml:"backend"    json:"backend"`
	Image     string          `yaml:"image"      json:"image"`
	Resources ResourcesConfig `yaml:"resources"  json:"resources"`
	Mode      string          `yaml:"mode"       json:"mode"`
	Che       CheConfig       `yaml:"che"        json:"che"`
	DemoPorts []int           `yaml:"demo_ports" json:"demo_ports"`
}

// ResourcesConfig defines container resource limits.
type ResourcesConfig struct {
	Memory string `yaml:"memory" json:"memory"`
	CPUs   string `yaml:"cpus"   json:"cpus"`
}

// CheConfig defines Eclipse Che / Dev Spaces settings.
type CheConfig struct {
	URL   string `yaml:"url"   json:"url"`
	Token string `yaml:"token" json:"token"`
}

// GatewayConfig controls `uf gateway` behavior.
type GatewayConfig struct {
	Port     int    `yaml:"port"     json:"port"`
	Provider string `yaml:"provider" json:"provider"`
}

// DoctorConfig controls `uf doctor` check behavior.
type DoctorConfig struct {
	Skip  []string          `yaml:"skip"  json:"skip"`
	Tools map[string]string `yaml:"tools" json:"tools"`
}

// WorkflowConfig controls hero lifecycle workflow. This section
// existed in the original .uf/config.yaml and is preserved here
// as one of seven sections.
type WorkflowConfig struct {
	ExecutionModes map[string]string `yaml:"execution_modes" json:"execution_modes"`
	SpecReview     bool              `yaml:"spec_review"     json:"spec_review"`
}

// LoadOptions controls how config files are located and read.
// All function fields default to production implementations
// when zero-valued.
type LoadOptions struct {
	ProjectDir    string
	ReadFile      func(string) ([]byte, error)
	Getenv        func(string) string
	UserConfigDir func() (string, error)
}

// defaults populates zero-value fields with production
// implementations.
func (o *LoadOptions) defaults() {
	if o.ReadFile == nil {
		o.ReadFile = os.ReadFile
	}
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.UserConfigDir == nil {
		o.UserConfigDir = os.UserConfigDir
	}
}

// Load reads the unified config with layered resolution:
//
//	compiled defaults → user config → repo config → env overrides
//
// Missing files are not errors — they produce compiled defaults.
// CLI flag overrides are applied by the caller at the cmd layer.
func Load(opts LoadOptions) (*Config, error) {
	opts.defaults()

	cfg := Defaults()

	// Layer 1: user-level config.
	userDir, err := opts.UserConfigDir()
	if err == nil {
		userPath := filepath.Join(userDir, "uf", "config.yaml")
		if data, readErr := opts.ReadFile(userPath); readErr == nil {
			var userCfg Config
			if parseErr := goyaml.Unmarshal(data, &userCfg); parseErr != nil {
				log.Warn("config file exists but failed to parse, using defaults",
					"path", userPath, "error", parseErr)
			} else {
				cfg = merge(cfg, userCfg)
			}
		}
	}

	// Layer 2: repo-level config.
	repoPath := filepath.Join(opts.ProjectDir, ".uf", "config.yaml")
	if data, readErr := opts.ReadFile(repoPath); readErr == nil {
		var repoCfg Config
		if parseErr := goyaml.Unmarshal(data, &repoCfg); parseErr != nil {
			log.Warn("config file exists but failed to parse, using defaults",
				"path", repoPath, "error", parseErr)
		} else {
			cfg = merge(cfg, repoCfg)
		}
	}

	// Layer 3: environment variable overrides.
	cfg = applyEnvOverrides(cfg, opts.Getenv)

	return &cfg, nil
}

// merge deep-merges overlay onto base. Non-zero values from
// overlay replace base. Slice fields are replaced (not appended).
// Map fields are merged key-by-key.
//
// Limitation: boolean fields use zero-value semantics, so
// `false` cannot override `true` from a lower-priority layer
// (e.g., setting `spec_review: false` in repo config will not
// override `spec_review: true` from user config). This is
// acceptable for the current schema where the only boolean
// (SpecReview) defaults to false. If more booleans are added
// with non-false defaults, consider using *bool pointer types.
func merge(base, overlay Config) Config {
	result := base

	// Setup
	if overlay.Setup.PackageManager != "" {
		result.Setup.PackageManager = overlay.Setup.PackageManager
	}
	if overlay.Setup.Skip != nil {
		result.Setup.Skip = overlay.Setup.Skip
	}
	if overlay.Setup.Tools != nil {
		if result.Setup.Tools == nil {
			result.Setup.Tools = make(map[string]ToolConfig)
		}
		for k, v := range overlay.Setup.Tools {
			result.Setup.Tools[k] = v
		}
	}

	// Scaffold
	if overlay.Scaffold.Language != "" {
		result.Scaffold.Language = overlay.Scaffold.Language
	}

	// Embedding
	if overlay.Embedding.Model != "" {
		result.Embedding.Model = overlay.Embedding.Model
	}
	if overlay.Embedding.Dimensions != 0 {
		result.Embedding.Dimensions = overlay.Embedding.Dimensions
	}
	if overlay.Embedding.Provider != "" {
		result.Embedding.Provider = overlay.Embedding.Provider
	}
	if overlay.Embedding.Host != "" {
		result.Embedding.Host = overlay.Embedding.Host
	}

	// Sandbox
	if overlay.Sandbox.Runtime != "" {
		result.Sandbox.Runtime = overlay.Sandbox.Runtime
	}
	if overlay.Sandbox.Backend != "" {
		result.Sandbox.Backend = overlay.Sandbox.Backend
	}
	if overlay.Sandbox.Image != "" {
		result.Sandbox.Image = overlay.Sandbox.Image
	}
	if overlay.Sandbox.Resources.Memory != "" {
		result.Sandbox.Resources.Memory = overlay.Sandbox.Resources.Memory
	}
	if overlay.Sandbox.Resources.CPUs != "" {
		result.Sandbox.Resources.CPUs = overlay.Sandbox.Resources.CPUs
	}
	if overlay.Sandbox.Mode != "" {
		result.Sandbox.Mode = overlay.Sandbox.Mode
	}
	if overlay.Sandbox.Che.URL != "" {
		result.Sandbox.Che.URL = overlay.Sandbox.Che.URL
	}
	if overlay.Sandbox.Che.Token != "" {
		result.Sandbox.Che.Token = overlay.Sandbox.Che.Token
	}
	if overlay.Sandbox.DemoPorts != nil {
		result.Sandbox.DemoPorts = overlay.Sandbox.DemoPorts
	}

	// Gateway
	if overlay.Gateway.Port != 0 {
		result.Gateway.Port = overlay.Gateway.Port
	}
	if overlay.Gateway.Provider != "" {
		result.Gateway.Provider = overlay.Gateway.Provider
	}

	// Doctor
	if overlay.Doctor.Skip != nil {
		result.Doctor.Skip = overlay.Doctor.Skip
	}
	if overlay.Doctor.Tools != nil {
		if result.Doctor.Tools == nil {
			result.Doctor.Tools = make(map[string]string)
		}
		for k, v := range overlay.Doctor.Tools {
			result.Doctor.Tools[k] = v
		}
	}

	// Workflow
	if overlay.Workflow.ExecutionModes != nil {
		if result.Workflow.ExecutionModes == nil {
			result.Workflow.ExecutionModes = make(map[string]string)
		}
		for k, v := range overlay.Workflow.ExecutionModes {
			result.Workflow.ExecutionModes[k] = v
		}
	}
	if overlay.Workflow.SpecReview {
		result.Workflow.SpecReview = true
	}

	return result
}

// applyEnvOverrides applies environment variable overrides to
// the config. Env vars have higher precedence than config files
// but lower than CLI flags.
func applyEnvOverrides(cfg Config, getenv func(string) string) Config {
	if v := getenv("UF_PACKAGE_MANAGER"); v != "" {
		cfg.Setup.PackageManager = v
	}
	if v := getenv("OLLAMA_MODEL"); v != "" {
		cfg.Embedding.Model = v
	}
	if v := getenv("OLLAMA_EMBED_DIM"); v != "" {
		if dim, err := strconv.Atoi(v); err == nil {
			cfg.Embedding.Dimensions = dim
		}
	}
	if v := getenv("OLLAMA_HOST"); v != "" {
		cfg.Embedding.Host = v
	}
	if v := getenv("UF_SANDBOX_IMAGE"); v != "" {
		cfg.Sandbox.Image = v
	}
	if v := getenv("UF_SANDBOX_BACKEND"); v != "" {
		cfg.Sandbox.Backend = v
	}
	if v := getenv("UF_SANDBOX_RUNTIME"); v != "" {
		cfg.Sandbox.Runtime = v
	}
	if v := getenv("UF_CHE_URL"); v != "" {
		cfg.Sandbox.Che.URL = v
	}
	if v := getenv("UF_CHE_TOKEN"); v != "" {
		cfg.Sandbox.Che.Token = v
	}
	if v := getenv("UF_GATEWAY_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Gateway.Port = port
		}
	}
	if v := getenv("UF_GATEWAY_PROVIDER"); v != "" {
		cfg.Gateway.Provider = v
	}
	return cfg
}

// Defaults returns a Config populated with all compiled defaults.
// These match the values currently hardcoded across setup.go,
// config.go, gateway.go, and checks.go.
func Defaults() Config {
	return Config{
		Setup: SetupConfig{
			PackageManager: "auto",
		},
		Scaffold: ScaffoldConfig{
			Language: "auto",
		},
		Embedding: EmbeddingConfig{
			Model:      "granite-embedding:30m",
			Dimensions: 256,
			Provider:   "ollama",
			Host:       "http://localhost:11434",
		},
		Sandbox: SandboxConfig{
			Runtime: "auto",
			Backend: "auto",
			Image:   "quay.io/unbound-force/opencode-dev:latest",
			Resources: ResourcesConfig{
				Memory: "8g",
				CPUs:   "4",
			},
			Mode: "isolated",
		},
		Gateway: GatewayConfig{
			Port:     53147,
			Provider: "auto",
		},
		Doctor: DoctorConfig{},
		Workflow: WorkflowConfig{
			ExecutionModes: map[string]string{
				"define":    "human",
				"implement": "swarm",
				"validate":  "swarm",
				"review":    "swarm",
				"accept":    "human",
				"reflect":   "swarm",
			},
			SpecReview: false,
		},
	}
}

// IsEmpty returns true if the sandbox section has no user-set
// values, indicating a fallback to .uf/sandbox.yaml is needed.
func (s SandboxConfig) IsEmpty() bool {
	d := Defaults().Sandbox
	return s.Runtime == d.Runtime &&
		s.Backend == d.Backend &&
		s.Image == d.Image &&
		s.Resources.Memory == d.Resources.Memory &&
		s.Resources.CPUs == d.Resources.CPUs &&
		s.Mode == d.Mode &&
		s.Che.URL == "" &&
		s.Che.Token == "" &&
		s.DemoPorts == nil
}

// RepoConfigPath returns the path to the repo-level config file.
func RepoConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ".uf", "config.yaml")
}

// UserConfigPath returns the path to the user-level config file,
// or an error if the user config directory cannot be determined.
func UserConfigPath(userConfigDir func() (string, error)) (string, error) {
	if userConfigDir == nil {
		userConfigDir = os.UserConfigDir
	}
	dir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, "uf", "config.yaml"), nil
}
