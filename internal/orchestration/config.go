package orchestration

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WorkflowConfig holds project-level workflow defaults read from
// .uf/config.yaml. Teams set this once to avoid passing
// CLI flags on every workflow start.
//
// Design decision: Struct uses nested anonymous struct for the
// "workflow" key to match the YAML structure directly, avoiding
// a separate type for a single-use inner struct (YAGNI).
type WorkflowConfig struct {
	Workflow struct {
		ExecutionModes map[string]string `yaml:"execution_modes"`
		SpecReview     bool              `yaml:"spec_review"`
	} `yaml:"workflow"`
}

// LoadWorkflowConfig reads .uf/config.yaml from the given
// directory. Returns a zero-value WorkflowConfig when the file
// does not exist (no error). Returns an error when the file
// exists but contains malformed YAML.
//
// The config path is filepath.Join(dir, "config.yaml"), matching
// the convention that dir is the .uf/ directory (same as
// WorkflowDir on the Orchestrator).
func LoadWorkflowConfig(dir string) (WorkflowConfig, error) {
	var cfg WorkflowConfig

	path := filepath.Join(dir, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Missing config file is not an error — return zero-value
			// defaults. Per Constitution Principle II (Composability
			// First): the config file is optional.
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
