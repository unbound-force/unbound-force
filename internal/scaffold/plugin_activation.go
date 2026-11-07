package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// pluginRuntimeRemediation and pluginInstallRemediation name the npm 11
	// escape hatch (npx -y npm@11.6.2): node 22's bundled npm 10.9.9
	// crashes in Arborist resolving the lockless peer graph when a target
	// keeps devDependencies, and the retry advice must not re-run the
	// crashing runtime. Doctor mirrors both strings (D2 twin commitment).
	pluginRuntimeRemediation = "install Node.js 22 and npm 11 (or run the install through npx -y npm@11.6.2), then rerun uf init"
	pluginInstallRemediation = "run npm install --ignore-scripts --omit=dev --no-package-lock in .opencode (through npm 11, e.g. npx -y npm@11.6.2, when npm 10 crashes resolving devDependencies), then rerun uf init"
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
	if len(activated) == 0 {
		return pluginActivationResult{result: subToolResult{
			name: "review-plugins", action: "skipped", detail: "opencode not found on PATH; npm dependencies installed",
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

	// No lock is shipped or written: .opencode/package.json pins the
	// direct dependencies exactly, and --no-package-lock keeps npm from
	// reading a stale lock or writing a new one that users and opencode
	// would then fight over.
	output, err = opts.ExecCmdInDir(opencodeDir, npmPath, "install", "--ignore-scripts", "--omit=dev", "--no-package-lock", "--no-audit", "--no-fund")
	if err != nil {
		return nil, output, fmt.Errorf("install plugin dependencies: %w; %s", err, pluginInstallRemediation)
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
		return nil, nil, nil
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
		ufWorkflowPluginAsset,
		ufFinalePluginAsset,
		"opencode/lib/agent-executor.ts",
		"opencode/lib/branch-classification.ts",
		"opencode/lib/model-matrix.ts",
		"opencode/lib/preflight-declared.ts",
		"opencode/lib/review-dispatch-artifact.ts",
		"opencode/lib/review-dispatch-lesson-proposal.ts",
		"opencode/lib/review-dispatch-sibling-evidence.ts",
		"opencode/lib/review-diff.ts",
		"opencode/lib/review-diff-manifest.ts",
		"opencode/lib/reviewer-manifest.ts",
		"opencode/lib/run-timeout.ts",
		"opencode/lib/uf-workflow-types.ts",
		"opencode/lib/unleash-templates.ts",
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
	workflowPath := filepath.ToSlash(filepath.Join(stageRoot, "plugins", "uf-workflow", "index.ts"))
	finalePath := filepath.ToSlash(filepath.Join(stageRoot, "plugins", "uf-finale", "index.ts"))
	probe := fmt.Sprintf(`import invokeAgentPlugin from %q
import reviewDispatchPlugin from %q
import ufWorkflowPlugin from %q
import ufFinalePlugin from %q

	`, pluginFileURL(invokePath), pluginFileURL(reviewPath), pluginFileURL(workflowPath), pluginFileURL(finalePath))
	if !toolDefinitions {
		return probe + "export default async () => ({})\n"
	}
	return probe + `export default async (input) => {
  const invokeHooks = await invokeAgentPlugin(input)
  const reviewHooks = await reviewDispatchPlugin(input)
  const workflowHooks = await ufWorkflowPlugin(input)
  const finaleHooks = await ufFinalePlugin(input)
  if (typeof invokeHooks.tool?.invoke_agent !== "object") {
    throw new Error("invoke-agent did not define invoke_agent")
  }
  for (const name of ["plan_review_dispatch", "finalize_review_dispatch", "acquire_sibling_evidence", "prepare_lesson_learning"]) {
    if (typeof reviewHooks.tool?.[name] !== "object") {
      throw new Error("review-dispatch did not define " + name)
    }
  }
  if (typeof workflowHooks.tool?.resolve_base_ref !== "object") {
    throw new Error("uf-workflow did not define resolve_base_ref")
  }
  if (typeof finaleHooks.tool?.uf_finale_startup !== "object") {
    throw new Error("uf-finale did not define uf_finale_startup")
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
		{asset: ufWorkflowPluginAsset, name: "uf-workflow"},
		{asset: ufFinalePluginAsset, name: "uf-finale"},
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
				// Copy the staged plugin source to the target location
				// using read+write instead of os.Rename for cross-filesystem
				// robustness (os.Rename of directories can fail with ENOENT
				// on some CI runner filesystem configurations).
				srcIndex := filepath.Join(source, "index.ts")
				content, readErr := opts.ReadFile(srcIndex)
				if readErr != nil {
					return nil, fmt.Errorf("read staged %s plugin source: %w", plugin.name, readErr)
				}
				if err := opts.RemoveAll(target); err != nil {
					return nil, fmt.Errorf("remove existing %s plugin directory: %w", plugin.name, err)
				}
				if err := opts.MkdirAll(target, 0o755); err != nil {
					return nil, fmt.Errorf("create %s plugin directory: %w", plugin.name, err)
				}
				destIndex := filepath.Join(target, "index.ts")
				if err := opts.WriteFile(destIndex, content, 0o644); err != nil {
					return nil, fmt.Errorf("write installed %s plugin source: %w", plugin.name, err)
				}
				if err := opts.RemoveAll(source); err != nil {
					return nil, fmt.Errorf("remove staged %s plugin source: %w", plugin.name, err)
				}
				activated = append(activated, mapAssetPath(plugin.asset))
				continue
			} else {
				if err := opts.RemoveAll(source); err != nil {
					return nil, fmt.Errorf("remove staged %s plugin source: %w", plugin.name, err)
				}
				activated = append(activated, mapAssetPath(plugin.asset))
				continue
			}
		}
		srcIndex := filepath.Join(source, "index.ts")
		content, readErr := opts.ReadFile(srcIndex)
		if readErr != nil {
			return nil, fmt.Errorf("read staged %s plugin source: %w", plugin.name, readErr)
		}
		if err := opts.MkdirAll(target, 0o755); err != nil {
			return nil, fmt.Errorf("create %s plugin directory: %w", plugin.name, err)
		}
		destIndex := filepath.Join(target, "index.ts")
		if err := opts.WriteFile(destIndex, content, 0o644); err != nil {
			return nil, fmt.Errorf("write installed %s plugin source: %w", plugin.name, err)
		}
		if err := opts.RemoveAll(source); err != nil {
			return nil, fmt.Errorf("remove staged %s plugin source: %w", plugin.name, err)
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

// pluginManifestAsset is the embedded path of the review plugin manifest.
const pluginManifestAsset = "opencode/package.json"

// preservePluginDevDependencies keeps a target's existing devDependencies
// when uf init refreshes the tool-owned .opencode/package.json. The shipped
// manifest has no devDependencies: activation installs with --omit=dev, yet
// a lockless npm install still resolves devDependencies, and the vitest 5
// peer graph crashes npm 10.9.9's Arborist. A repository that runs the
// plugin suite itself (this one) declares its devDependencies locally, and
// uf init must not strip them. The block is re-inserted before "overrides"
// with sorted keys in the manifest's two-space layout, so an unchanged
// target stays byte-identical and is skipped.
func preservePluginDevDependencies(shipped, existing []byte) []byte {
	var target struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(existing, &target); err != nil || len(target.DevDependencies) == 0 {
		return shipped
	}
	anchor := []byte("\n  \"overrides\": {\n")
	at := bytes.Index(shipped, anchor)
	if at < 0 {
		// A missing anchor is a silent passthrough by design: erroring
		// would fail uf init on a cosmetic manifest reformat, and
		// internal/scaffold deliberately has no logging or warnings
		// channel. The silence is contained at origin — the round-trip
		// test TestPreservePluginDevDependencies_RestoresRepositoryManifest
		// fails if the shipped manifest's overrides anchor changes
		// shape, and the anchor-miss case in the same file pins this
		// unchanged return.
		return shipped
	}
	names := make([]string, 0, len(target.DevDependencies))
	for name := range target.DevDependencies {
		names = append(names, name)
	}
	sort.Strings(names)
	var block bytes.Buffer
	block.WriteString("  \"devDependencies\": {\n")
	for i, name := range names {
		sep := ","
		if i == len(names)-1 {
			sep = ""
		}
		fmt.Fprintf(&block, "    %s: %s%s\n", jsonString(name), jsonString(target.DevDependencies[name]), sep)
	}
	block.WriteString("  },")
	out := make([]byte, 0, len(shipped)+block.Len()+1)
	out = append(out, shipped[:at+1]...)
	out = append(out, block.Bytes()...)
	out = append(out, shipped[at:]...)
	return out
}

// jsonString encodes s as a JSON string literal without HTML escaping, so
// version ranges such as ">=1.0.0" keep their literal form.
func jsonString(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	// Best-effort: encoding to an in-memory bytes.Buffer cannot fail.
	_ = enc.Encode(s)
	return strings.TrimSuffix(buf.String(), "\n")
}
