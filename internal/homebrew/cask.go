// Package homebrew provides release-publishing helpers for Homebrew assets.
package homebrew

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const legacyPostflightHook = `  postflight do
    if OS.mac?
      system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/unbound-force"]
      # Create uf symlink alias for daily-use convenience (FR-002/FR-009).
      system_command "/bin/ln", args: ["-sf", "#{staged_path}/unbound-force", "#{HOMEBREW_PREFIX}/bin/uf"]
    end
  end`

const declarativePostflightSteps = `  postflight_steps do
    on_macos do
      run "/usr/bin/xattr",
        args: ["-dr", "com.apple.quarantine", "{{staged_path}}/unbound-force"]
      symlink "unbound-force", "{{HOMEBREW_PREFIX}}/bin/uf", overwrite: true
    end
  end`

// Options configures a generated Cask transformation.
type Options struct {
	// InputPath is the generated Cask to transform.
	InputPath string

	// OutputPath is where Run writes the transformed Cask.
	OutputPath string

	// ReadFile reads the generated Cask. It defaults to os.ReadFile.
	ReadFile func(string) ([]byte, error)

	// WriteFile writes the transformed Cask. It defaults to os.WriteFile.
	WriteFile func(string, []byte, os.FileMode) error
}

// Result describes the transformed Cask written by Run.
type Result struct {
	// OutputPath is the location of the transformed Cask.
	OutputPath string
}

// Run transforms the one known legacy postflight hook in a generated Cask.
// It returns an error when the exact hook is absent or ambiguous.
func Run(options Options) (*Result, error) {
	if options.InputPath == "" {
		return nil, fmt.Errorf("transform Homebrew Cask hook: input path is required")
	}
	if options.OutputPath == "" {
		return nil, fmt.Errorf("transform Homebrew Cask hook: output path is required")
	}
	if options.ReadFile == nil {
		options.ReadFile = os.ReadFile
	}
	if options.WriteFile == nil {
		options.WriteFile = os.WriteFile
	}

	cask, err := options.ReadFile(options.InputPath)
	if err != nil {
		return nil, fmt.Errorf("read generated Homebrew Cask: %w", err)
	}

	transformed, err := Transform(string(cask))
	if err != nil {
		return nil, err
	}
	temporaryFile, err := os.CreateTemp(filepath.Dir(options.OutputPath), "."+filepath.Base(options.OutputPath)+"-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary transformed Homebrew Cask: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer func() {
		_ = os.Remove(temporaryPath)
	}()
	if err := temporaryFile.Close(); err != nil {
		return nil, fmt.Errorf("close temporary transformed Homebrew Cask: %w", err)
	}

	if err := options.WriteFile(temporaryPath, []byte(transformed), 0o644); err != nil {
		return nil, fmt.Errorf("write transformed Homebrew Cask: %w", err)
	}
	if err := os.Chmod(temporaryPath, 0o644); err != nil {
		return nil, fmt.Errorf("set transformed Homebrew Cask permissions: %w", err)
	}
	if err := os.Rename(temporaryPath, options.OutputPath); err != nil {
		return nil, fmt.Errorf("replace transformed Homebrew Cask: %w", err)
	}

	return &Result{OutputPath: options.OutputPath}, nil
}

// Transform replaces exactly one literal legacy postflight hook with the
// supported declarative postflight_steps block.
func Transform(cask string) (string, error) {
	occurrences := strings.Count(cask, legacyPostflightHook)
	if occurrences != 1 {
		return "", fmt.Errorf("transform generated Homebrew Cask hook: expected exactly one legacy postflight hook, found %d", occurrences)
	}

	return strings.Replace(cask, legacyPostflightHook, declarativePostflightSteps, 1), nil
}
