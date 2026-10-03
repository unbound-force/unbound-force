package scaffold

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	pluginRuntimeRemediation = "install Node.js 22 and npm 10, then rerun uf init"
	pluginInstallRemediation = "run npm ci --ignore-scripts --omit=dev in .opencode, then rerun uf init"
	pluginProbeDirectoryName = "uf-review-plugin-probe"
)

type runtimeVersion struct {
	major uint32
	minor uint32
	patch uint32
}

type pluginActivationResult struct {
	result    subToolResult
	activated []string
}

func activateReviewPlugins(opts *Options) pluginActivationResult {
	normalizePluginActivationOptions(opts)
	if opts.DryRun {
		return pluginActivationResult{result: subToolResult{
			name: "review-plugins", action: "dry-run", detail: "activation: inactive",
		}}
	}

	activated, output, err := installAndActivateReviewPlugins(opts)
	if err != nil {
		return pluginActivationResult{result: subToolResult{
			name:   "review-plugins",
			action: "failed",
			detail: fmt.Sprintf("activation: inactive; %v", err),
			err:    err,
			output: output,
		}}
	}

	return pluginActivationResult{
		result: subToolResult{
			name: "review-plugins", action: "activated", detail: "activation: active",
		},
		activated: activated,
	}
}

func installAndActivateReviewPlugins(opts *Options) ([]string, []byte, error) {
	opencodeDir := filepath.Join(opts.TargetDir, ".opencode")
	npmPath, output, err := validatePluginRuntime(opts, opencodeDir)
	if err != nil {
		return nil, output, err
	}

	output, err = opts.ExecCmdInDir(opencodeDir, npmPath, "ci", "--ignore-scripts", "--omit=dev")
	if err != nil {
		return nil, output, fmt.Errorf("install locked plugin dependencies: %w; %s", err, pluginInstallRemediation)
	}

	stageRoot, err := opts.MkdirTemp(opencodeDir, ".uf-review-plugins-")
	if err != nil {
		return nil, nil, fmt.Errorf("create external plugin stage: %w", err)
	}
	defer func() { _ = opts.RemoveAll(stageRoot) }()
	if err := stageReviewPluginSources(opts, stageRoot); err != nil {
		return nil, nil, err
	}

	opencodePath, err := opts.LookPath("opencode")
	if err != nil {
		return nil, nil, fmt.Errorf("resolve opencode for provider-free plugin probes: %w", err)
	}
	probeDirectory := filepath.Join(opencodeDir, "plugins", pluginProbeDirectoryName)
	defer func() { _ = opts.RemoveAll(probeDirectory) }()
	if err := runReviewPluginProbe(opts, opencodeDir, opencodePath, probeDirectory, stageRoot, false); err != nil {
		return nil, nil, fmt.Errorf("probe staged plugin imports: %w", err)
	}
	if err := runReviewPluginProbe(opts, opencodeDir, opencodePath, probeDirectory, stageRoot, true); err != nil {
		return nil, nil, fmt.Errorf("probe staged plugin tool definitions: %w", err)
	}
	if err := opts.RemoveAll(probeDirectory); err != nil {
		return nil, nil, fmt.Errorf("remove provider-free plugin probe: %w", err)
	}

	activated, err := atomicallyActivateReviewPlugins(opts, stageRoot)
	if err != nil {
		return nil, nil, err
	}
	if err := opts.RemoveAll(stageRoot); err != nil {
		return nil, nil, fmt.Errorf("remove external plugin stage: %w", err)
	}
	return activated, nil, nil
}

func validatePluginRuntime(opts *Options, workingDir string) (string, []byte, error) {
	nodePath, err := opts.LookPath("node")
	if err != nil {
		return "", nil, fmt.Errorf("resolve Node.js: %w; %s", err, pluginRuntimeRemediation)
	}
	nodeOutput, err := opts.ExecCmdInDir(workingDir, nodePath, "--version")
	if err != nil {
		return "", nodeOutput, fmt.Errorf("read Node.js version (detected %q): %w; %s",
			string(nodeOutput), err, pluginRuntimeRemediation)
	}
	nodeVersion, err := parseRuntimeVersion(nodeOutput, true)
	if err != nil {
		return "", nodeOutput, fmt.Errorf("malformed Node.js version (detected %q): %w; %s",
			string(nodeOutput), err, pluginRuntimeRemediation)
	}
	if nodeVersion.major < 20 || nodeVersion.major > 24 {
		return "", nodeOutput, fmt.Errorf("unsupported Node.js version (detected %q): require Node 20 through 24; %s",
			string(nodeOutput), pluginRuntimeRemediation)
	}

	npmPath, err := opts.LookPath("npm")
	if err != nil {
		return "", nil, fmt.Errorf("resolve npm: %w; %s", err, pluginRuntimeRemediation)
	}
	npmOutput, err := opts.ExecCmdInDir(workingDir, npmPath, "--version")
	if err != nil {
		return "", npmOutput, fmt.Errorf("read npm version (detected %q): %w; %s",
			string(npmOutput), err, pluginRuntimeRemediation)
	}
	npmVersion, err := parseRuntimeVersion(npmOutput, false)
	if err != nil {
		return "", npmOutput, fmt.Errorf("malformed npm version (detected %q): %w; %s",
			string(npmOutput), err, pluginRuntimeRemediation)
	}
	if npmVersion.major < 10 || npmVersion.major > 11 {
		return "", npmOutput, fmt.Errorf("unsupported npm version (detected %q): require npm 10 through 11; %s",
			string(npmOutput), pluginRuntimeRemediation)
	}
	return npmPath, nil, nil
}

func parseRuntimeVersion(output []byte, allowVPrefix bool) (runtimeVersion, error) {
	value := string(output)
	if strings.HasSuffix(value, "\r\n") {
		value = strings.TrimSuffix(value, "\r\n")
	} else if strings.HasSuffix(value, "\n") {
		value = strings.TrimSuffix(value, "\n")
	}
	value = trimASCIISpaceAndTab(value)
	if allowVPrefix && strings.HasPrefix(value, "v") {
		value = value[1:]
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return runtimeVersion{}, fmt.Errorf("version must contain exactly three ASCII decimal components")
	}

	components := [3]uint32{}
	for index, part := range parts {
		component, err := parseUint32VersionComponent(part)
		if err != nil {
			return runtimeVersion{}, fmt.Errorf("version component %d: %w", index+1, err)
		}
		components[index] = component
	}
	return runtimeVersion{major: components[0], minor: components[1], patch: components[2]}, nil
}

func trimASCIISpaceAndTab(value string) string {
	return strings.Trim(value, " \t")
}

func parseUint32VersionComponent(value string) (uint32, error) {
	if value == "" {
		return 0, fmt.Errorf("component is empty")
	}
	if len(value) > 1 && value[0] == '0' {
		return 0, fmt.Errorf("leading zeroes are not allowed")
	}
	if len(value) > 10 {
		return 0, fmt.Errorf("component exceeds uint32")
	}

	var parsed uint32
	for _, character := range []byte(value) {
		if character < '0' || character > '9' {
			return 0, fmt.Errorf("component must contain ASCII digits only")
		}
		digit := uint32(character - '0')
		if parsed > (^uint32(0)-digit)/10 {
			return 0, fmt.Errorf("component exceeds uint32")
		}
		parsed = parsed*10 + digit
	}
	return parsed, nil
}

func stageReviewPluginSources(opts *Options, stageRoot string) error {
	assetPaths := []string{
		invokeAgentPluginAsset,
		reviewDispatchPluginAsset,
		"opencode/lib/review-dispatch-lesson-proposal.ts",
		"opencode/lib/review-dispatch-sibling-evidence.ts",
		"opencode/lib/reviewer-manifest.ts",
	}
	for _, assetPath := range assetPaths {
		content, err := assetContent(assetPath)
		if err != nil {
			return fmt.Errorf("read staged plugin asset %s: %w", assetPath, err)
		}
		relativePath := strings.TrimPrefix(assetPath, "opencode/")
		targetPath := filepath.Join(stageRoot, filepath.FromSlash(relativePath))
		if err := opts.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create staged plugin directory: %w", err)
		}
		if err := opts.WriteFile(targetPath, content, 0o644); err != nil {
			return fmt.Errorf("write staged plugin asset %s: %w", assetPath, err)
		}
	}
	return nil
}

func runReviewPluginProbe(
	opts *Options,
	workingDir, opencodePath, probeDirectory, stageRoot string,
	toolDefinitions bool,
) error {
	if err := opts.MkdirAll(probeDirectory, 0o755); err != nil {
		return fmt.Errorf("create provider-free probe directory: %w", err)
	}
	probePath := filepath.Join(probeDirectory, "index.ts")
	if err := opts.WriteFile(probePath, []byte(reviewPluginProbeSource(stageRoot, toolDefinitions)), 0o644); err != nil {
		return fmt.Errorf("write provider-free probe: %w", err)
	}
	output, err := opts.ExecCmdInDir(workingDir, opencodePath, "debug", "config")
	if err != nil {
		return fmt.Errorf("opencode debug config: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func reviewPluginProbeSource(stageRoot string, toolDefinitions bool) string {
	invokePath := filepath.ToSlash(filepath.Join(stageRoot, "plugins", "invoke-agent", "index.ts"))
	reviewPath := filepath.ToSlash(filepath.Join(stageRoot, "plugins", "review-dispatch", "index.ts"))
	probe := fmt.Sprintf(`import invokeAgentPlugin from %q
import reviewDispatchPlugin from %q

	`, pluginFileURL(invokePath), pluginFileURL(reviewPath))
	if !toolDefinitions {
		return probe + "export default async () => ({})\n"
	}
	return probe + `export default async (input) => {
  const invokeHooks = await invokeAgentPlugin(input)
  const reviewHooks = await reviewDispatchPlugin(input)
  if (typeof invokeHooks.tool?.invoke_agent !== "object") {
    throw new Error("invoke-agent did not define invoke_agent")
  }
  for (const name of ["plan_review_dispatch", "finalize_review_dispatch", "acquire_sibling_evidence", "prepare_lesson_learning"]) {
    if (typeof reviewHooks.tool?.[name] !== "object") {
      throw new Error("review-dispatch did not define " + name)
    }
  }
  return {}
}`
}

func pluginFileURL(path string) string {
	slashPath := filepath.ToSlash(path)
	if !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}
	return (&url.URL{Scheme: "file", Path: slashPath}).String()
}

func atomicallyActivateReviewPlugins(opts *Options, stageRoot string) ([]string, error) {
	pluginsDirectory := filepath.Join(opts.TargetDir, ".opencode", "plugins")
	if err := opts.MkdirAll(pluginsDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("create plugin auto-discovery directory: %w", err)
	}

	assetsToActivate := []struct {
		asset string
		name  string
	}{
		{asset: invokeAgentPluginAsset, name: "invoke-agent"},
		{asset: reviewDispatchPluginAsset, name: "review-dispatch"},
	}
	activated := make([]string, 0, len(assetsToActivate))
	for _, plugin := range assetsToActivate {
		source := filepath.Join(stageRoot, "plugins", plugin.name)
		target := filepath.Join(pluginsDirectory, plugin.name)
		if _, statErr := opts.Stat(target); statErr == nil {
			// Check if target is a symlink - if so, always refresh for security
			var needsRefresh bool
			if linfo, lstatErr := opts.Lstat(target); lstatErr == nil && (linfo.Mode()&os.ModeSymlink != 0) {
				needsRefresh = true
			} else if opts.Force {
				needsRefresh = true
			} else {
				// Read staged content - if this fails, we should error out for security
				stagedContent, readErr := opts.ReadFile(filepath.Join(source, "index.ts"))
				if readErr != nil {
					return nil, fmt.Errorf("read staged %s plugin source: %w", plugin.name, readErr)
				}
				// Read installed content - if this fails, refresh for safety
				installedContent, readErr2 := opts.ReadFile(filepath.Join(target, "index.ts"))
				needsRefresh = readErr2 != nil || !bytes.Equal(stagedContent, installedContent)
			}
			if needsRefresh {
				// More atomic refresh: backup existing, move staged, restore on failure
				backup := target + ".backup"
				var backupCreated bool
				
				// Create backup of existing target if it exists
				if _, err := opts.Stat(target); err == nil {
					if err := opts.Rename(target, backup); err != nil {
						return nil, fmt.Errorf("backup existing %s plugin: %w", plugin.name, err)
					}
					backupCreated = true
				}
				
				// Try to move staged plugin to target location
				if err := opts.Rename(source, target); err != nil {
					// If rename failed and we have a backup, try to restore it
					if backupCreated {
						if restoreErr := opts.Rename(backup, target); restoreErr != nil {
							return nil, fmt.Errorf("failed to activate %s plugin and failed to restore backup: %w (original error: %v)", plugin.name, restoreErr, err)
						}
						// Successfully restored, return the original error
						return nil, fmt.Errorf("atomically activate %s plugin source: %w", plugin.name, err)
					}
					// No backup to restore, just return the error
					return nil, fmt.Errorf("atomically activate %s plugin source: %w", plugin.name, err)
				}
				
				// If we successfully moved the plugin, clean up the backup
				if backupCreated {
					if err := opts.RemoveAll(backup); err != nil {
						// Log but don't fail - the activation was successful
						// This is a cleanup error, not a critical failure
						// We could log this with a proper logger, but for now we'll just ignore it
						_ = err // Explicitly ignore the error to satisfy the linter
					}
				}
			} else {
				if err := opts.RemoveAll(source); err != nil {
					return nil, fmt.Errorf("remove staged %s plugin source: %w", plugin.name, err)
				}
				activated = append(activated, mapAssetPath(plugin.asset))
				continue
			}
		}
		if err := opts.Rename(source, target); err != nil {
			return nil, fmt.Errorf("atomically activate %s plugin source: %w", plugin.name, err)
		}
		activated = append(activated, mapAssetPath(plugin.asset))
	}
	return activated, nil
}

// normalizePluginActivationOptions supplies production dependencies for direct
// activation callers that do not enter through Run.
func normalizePluginActivationOptions(opts *Options) {
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.ExecCmdInDir == nil {
		opts.ExecCmdInDir = defaultExecCmdInDir
	}
	if opts.ReadFile == nil {
		opts.ReadFile = os.ReadFile
	}
	if opts.Stat == nil {
		opts.Stat = os.Stat
	}
	if opts.Lstat == nil {
		opts.Lstat = os.Lstat
	}
	if opts.WriteFile == nil {
		opts.WriteFile = os.WriteFile
	}
	if opts.MkdirTemp == nil {
		opts.MkdirTemp = os.MkdirTemp
	}
	if opts.MkdirAll == nil {
		opts.MkdirAll = os.MkdirAll
	}
	if opts.Rename == nil {
		opts.Rename = os.Rename
	}
	if opts.RemoveAll == nil {
		opts.RemoveAll = os.RemoveAll
	}
}
