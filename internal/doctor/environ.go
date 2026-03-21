package doctor

import (
	"runtime"
	"strings"
)

// DetectEnvironment discovers the developer's version and package
// managers by checking PATH presence and environment variables.
// Detection uses the injected LookPath, EvalSymlinks, and Getenv
// from Options for testability per Constitution Principle IV.
func DetectEnvironment(opts *Options) DetectedEnvironment {
	env := DetectedEnvironment{
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}

	// Detection order follows research.md R1: most specific first.
	// Each manager is checked independently — multiple can coexist.

	// goenv: Go version manager (shim-based)
	if path, err := opts.LookPath("goenv"); err == nil {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerGoenv,
			Path:    path,
			Manages: []string{"go"},
		})
	} else if opts.Getenv("GOENV_ROOT") != "" {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerGoenv,
			Path:    opts.Getenv("GOENV_ROOT"),
			Manages: []string{"go"},
		})
	}

	// pyenv: Python version manager (shim-based)
	if path, err := opts.LookPath("pyenv"); err == nil {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerPyenv,
			Path:    path,
			Manages: []string{"python"},
		})
	} else if opts.Getenv("PYENV_ROOT") != "" {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerPyenv,
			Path:    opts.Getenv("PYENV_ROOT"),
			Manages: []string{"python"},
		})
	}

	// nvm: Node Version Manager (bash function, not a binary).
	// Detect by NVM_DIR env var since nvm is a shell function.
	if nvmDir := opts.Getenv("NVM_DIR"); nvmDir != "" {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerNvm,
			Path:    nvmDir,
			Manages: []string{"node"},
		})
	}

	// fnm: Fast Node Manager (binary, multishell)
	if path, err := opts.LookPath("fnm"); err == nil {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerFnm,
			Path:    path,
			Manages: []string{"node"},
		})
	} else if opts.Getenv("FNM_DIR") != "" {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerFnm,
			Path:    opts.Getenv("FNM_DIR"),
			Manages: []string{"node"},
		})
	}

	// mise: Polyglot version manager (formerly rtx)
	if path, err := opts.LookPath("mise"); err == nil {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerMise,
			Path:    path,
			Manages: []string{"go", "node", "python"},
		})
	}

	// bun: Bun JavaScript runtime
	if path, err := opts.LookPath("bun"); err == nil {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerBun,
			Path:    path,
			Manages: []string{"node", "packages"},
		})
	}

	// Homebrew: macOS/Linux package manager
	if path, err := opts.LookPath("brew"); err == nil {
		env.Managers = append(env.Managers, ManagerInfo{
			Kind:    ManagerHomebrew,
			Path:    path,
			Manages: []string{"packages"},
		})
	}

	// Ensure Managers is empty slice, not nil, per data-model.md.
	if env.Managers == nil {
		env.Managers = []ManagerInfo{}
	}

	return env
}

// DetectProvenance determines which manager installed a binary
// at the given path using the 10-step priority chain from
// research.md R1. Uses injected EvalSymlinks for Homebrew
// symlink resolution.
func DetectProvenance(binaryPath string, opts *Options) ManagerKind {
	if binaryPath == "" {
		return ManagerUnknown
	}

	// 1. goenv — path contains /.goenv/shims/ or /.goenv/versions/
	if strings.Contains(binaryPath, "/.goenv/shims/") ||
		strings.Contains(binaryPath, "/.goenv/versions/") {
		return ManagerGoenv
	}

	// 2. pyenv — path contains /.pyenv/shims/ or /.pyenv/versions/
	if strings.Contains(binaryPath, "/.pyenv/shims/") ||
		strings.Contains(binaryPath, "/.pyenv/versions/") {
		return ManagerPyenv
	}

	// 3. nvm — path contains /.nvm/versions/
	if strings.Contains(binaryPath, "/.nvm/versions/") {
		return ManagerNvm
	}

	// 4. fnm — path contains /fnm_multishells/ or /fnm/node-versions/
	if strings.Contains(binaryPath, "/fnm_multishells/") ||
		strings.Contains(binaryPath, "/fnm/node-versions/") {
		return ManagerFnm
	}

	// 5. mise — path contains /mise/installs/ or /mise/shims/
	if strings.Contains(binaryPath, "/mise/installs/") ||
		strings.Contains(binaryPath, "/mise/shims/") {
		return ManagerMise
	}

	// 6. bun — path contains /.bun/bin/
	if strings.Contains(binaryPath, "/.bun/bin/") {
		return ManagerBun
	}

	// 7. Homebrew — resolved symlink path contains /Cellar/
	// Critical: must resolve symlinks for Homebrew detection
	// on Intel macOS where /usr/local/bin is ambiguous.
	if opts.EvalSymlinks != nil {
		if resolved, err := opts.EvalSymlinks(binaryPath); err == nil {
			if strings.Contains(resolved, "/Cellar/") {
				return ManagerHomebrew
			}
		}
	}

	// 8. Direct install — path starts with /usr/local/go/bin/
	if strings.HasPrefix(binaryPath, "/usr/local/go/bin/") {
		return ManagerDirect
	}

	// 9. System — path starts with /usr/bin/ or /snap/bin/
	if strings.HasPrefix(binaryPath, "/usr/bin/") ||
		strings.HasPrefix(binaryPath, "/snap/bin/") {
		return ManagerSystem
	}

	// 10. Unknown — no match
	return ManagerUnknown
}

// installHint returns a manager-appropriate install command for
// a given tool name based on the detected environment. Falls back
// to generic instructions when no specific manager is detected.
func installHint(toolName string, env DetectedEnvironment) string {
	// Check if a specific manager is available for this tool.
	for _, m := range env.Managers {
		for _, managed := range m.Manages {
			if managed == toolCategory(toolName) {
				return managerInstallCmd(toolName, m.Kind)
			}
		}
	}

	// Check if Homebrew is available as a fallback.
	for _, m := range env.Managers {
		if m.Kind == ManagerHomebrew {
			return homebrewInstallCmd(toolName)
		}
	}

	// No manager detected — return generic instructions.
	return genericInstallCmd(toolName)
}

// toolCategory maps a tool binary name to its manager category.
func toolCategory(toolName string) string {
	switch toolName {
	case "go":
		return "go"
	case "node", "npm":
		return "node"
	case "python", "python3":
		return "python"
	default:
		return "packages"
	}
}

// managerInstallCmd returns the install command for a tool via
// a specific manager.
func managerInstallCmd(toolName string, manager ManagerKind) string {
	switch manager {
	case ManagerGoenv:
		switch toolName {
		case "go":
			return "goenv install 1.24.3 && goenv global 1.24.3"
		}
	case ManagerNvm:
		switch toolName {
		case "node":
			return "nvm install 22"
		}
	case ManagerFnm:
		switch toolName {
		case "node":
			return "fnm install 22"
		}
	case ManagerMise:
		switch toolName {
		case "go":
			return "mise install go@1.24"
		case "node":
			return "mise install node@22"
		}
	case ManagerHomebrew:
		return homebrewInstallCmd(toolName)
	case ManagerBun:
		switch toolName {
		case "swarm":
			return "bun add -g opencode-swarm-plugin@latest"
		}
	}
	return homebrewInstallCmd(toolName)
}

// homebrewInstallCmd returns the Homebrew install command for a tool.
func homebrewInstallCmd(toolName string) string {
	switch toolName {
	case "go":
		return "brew install go"
	case "opencode":
		return "brew install anomalyco/tap/opencode"
	case "gaze":
		return "brew install unbound-force/tap/gaze"
	case "mxf":
		return "brew install unbound-force/tap/mxf"
	case "graphthulhu":
		return "brew install unbound-force/tap/graphthulhu"
	case "node":
		return "brew install node"
	case "gh":
		return "brew install gh"
	case "swarm":
		return "npm install -g opencode-swarm-plugin@latest"
	default:
		return "brew install " + toolName
	}
}

// genericInstallCmd returns generic install instructions when
// no package manager is detected.
func genericInstallCmd(toolName string) string {
	switch toolName {
	case "opencode":
		return "curl -fsSL https://opencode.ai/install | bash"
	case "gaze":
		return "Download from https://github.com/unbound-force/gaze/releases"
	case "go":
		return "Download from https://go.dev/dl/"
	case "node":
		return "Download from https://nodejs.org/"
	case "swarm":
		return "npm install -g opencode-swarm-plugin@latest"
	case "gh":
		return "Download from https://cli.github.com/"
	default:
		return "Install " + toolName
	}
}

// HasManager checks if a specific manager kind is present in
// the detected environment.
func HasManager(env DetectedEnvironment, kind ManagerKind) bool {
	for _, m := range env.Managers {
		if m.Kind == kind {
			return true
		}
	}
	return false
}

// installURL returns a documentation URL for non-trivial installs.
func installURL(toolName string) string {
	switch toolName {
	case "opencode":
		return "https://opencode.ai/docs"
	case "gaze":
		return "https://github.com/unbound-force/gaze"
	case "node":
		return "https://nodejs.org/"
	case "go":
		return "https://go.dev/dl/"
	case "gh":
		return "https://cli.github.com/"
	case "graphthulhu":
		return "https://github.com/skridlevsky/graphthulhu"
	default:
		return ""
	}
}
