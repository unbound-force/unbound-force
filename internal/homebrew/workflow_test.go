package homebrew

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// The release pipeline is delegated to jflowers/ship's shared go-release.yml
// (the same pipeline replicator, dewey and gaze use). The tests below pin
// the delegation contract: the immutable workflow pin, the inputs and
// secrets mapped into it, the job permissions it runs with, and the
// GoReleaser cask block the pipeline publishes. The publication ordering
// and fail-closed validation logic itself lives in the pinned ship
// workflow, so an immutable commit SHA (never a mutable tag) is what keeps
// it trustworthy.

// immutableCommitSHA matches a 40-character commit SHA. The release
// pipeline pin must be commit-addressed; mutable name-addressed refs
// (tags, branches) are forbidden.
var immutableCommitSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// goreleaserEscapedTemplates maps GoReleaser's backtick-escaped Homebrew
// templates to the literal text the release cask must carry. GoReleaser
// renders {{ `X` }} as X, keeping Homebrew's own {{...}} templates
// unexpanded; the transformer writes the resolved form directly.
var goreleaserEscapedTemplates = map[string]string{
	"{{ `{{staged_path}}` }}":     "{{staged_path}}",
	"{{ `{{HOMEBREW_PREFIX}}` }}": "{{HOMEBREW_PREFIX}}",
}

func TestReleaseWorkflow_DelegatesToPinnedShipPipeline(t *testing.T) {
	t.Parallel()

	workflow := releaseWorkflow(t)

	const shipPipeline = "uses: jflowers/ship/.github/workflows/go-release.yml@"
	usesLine := lineContaining(workflow, shipPipeline)
	if usesLine == "" {
		t.Fatal("release workflow must delegate to jflowers/ship's go-release.yml")
	}
	pin := strings.TrimSpace(usesLine[strings.Index(usesLine, "@")+1:])
	if comment := strings.Index(pin, "#"); comment != -1 {
		pin = strings.TrimSpace(pin[:comment])
	}
	if !immutableCommitSHA.MatchString(pin) {
		t.Fatalf("ship pipeline must be pinned to an immutable commit SHA, got %q", pin)
	}

	for _, input := range []string{
		"tag: ${{ inputs.tag }}",
		"allow_prerelease: true",
		"skip_semver_check: ${{ inputs.skip_semver_check }}",
		"skip_ci_checks: ${{ inputs.skip_ci_checks }}",
		"skip_unreleased_check: ${{ inputs.skip_unreleased_check }}",
		"tap: unbound-force/homebrew-tap",
		"formula: unbound-force",
	} {
		if !strings.Contains(workflow, input) {
			t.Errorf("release workflow is missing ship input %q", input)
		}
	}

	ciChecks := lineContaining(workflow, "      ci_checks:")
	for _, check := range []string{
		"Build and Test",
		"Standardized CI / Run linters",
		"OSV-Scanner / Trivy Source Scan",
		"Plugin Tests",
	} {
		if !strings.Contains(ciChecks, check) {
			t.Errorf("release ci_checks must gate on %q", check)
		}
	}

	for _, check := range []string{
		`test -x "$HOMEBREW_PREFIX/bin/unbound-force"`,
		`test -L "$HOMEBREW_PREFIX/bin/uf"`,
		`"$HOMEBREW_PREFIX/bin/uf" --version`,
	} {
		if !strings.Contains(workflow, check) {
			t.Errorf("release cask_install_check is missing %q", check)
		}
	}

	for _, secret := range []string{
		"MACOS_SIGN_P12",
		"MACOS_SIGN_PASSWORD",
		"MACOS_SIGN_IDENTITY",
		"MACOS_NOTARY_KEY",
		"MACOS_NOTARY_KEY_ID",
		"MACOS_NOTARY_ISSUER_ID",
		"HOMEBREW_TAP_TOKEN",
	} {
		mapping := secret + ": ${{ secrets." + secret + " }}"
		if !strings.Contains(workflow, mapping) {
			t.Errorf("release workflow is missing secret mapping %q", mapping)
		}
	}

	if !strings.Contains(workflow, "permissions: {}") {
		t.Error("release workflow must default to no permissions (permissions: {})")
	}
	grants := jobPermissionGrants(workflow)
	sort.Strings(grants)
	const wantGrants = "checks: read\ncontents: write\nid-token: write"
	if got := strings.Join(grants, "\n"); got != wantGrants {
		t.Errorf("release job permissions = %q, want exactly %q", got, wantGrants)
	}

	if !strings.Contains(workflow, "on:\n  workflow_dispatch:") {
		t.Error("release workflow must trigger only via workflow_dispatch")
	}
	if !strings.Contains(workflow, "cancel-in-progress: false") {
		t.Error("release workflow must not cancel in-progress releases")
	}
	if strings.Contains(workflow, "transform-homebrew-cask") {
		t.Error("release workflow must not run transform-homebrew-cask: the cask's postflight_steps are written by GoReleaser's custom_block")
	}
}

func TestReleaseWorkflow_GoReleaserCaskMatchesTransformerOutput(t *testing.T) {
	t.Parallel()

	manifest := goreleaserManifest(t)

	block := goreleaserCustomBlock(t, manifest)
	if strings.Contains(block, "postflight do") {
		t.Error("GoReleaser custom_block must not render the deprecated postflight do hook")
	}
	got := dedentBlock(unescapeGoReleaserTemplates(block))
	want := dedentBlock(declarativePostflightSteps)
	if got != want {
		t.Errorf("GoReleaser custom_block must match the transformer's declarative postflight_steps\ngot:\n%s\nwant:\n%s", got, want)
	}
	skipUploads := 0
	for _, line := range strings.Split(manifest, "\n") {
		if strings.TrimSpace(line) == "skip_upload: true" {
			skipUploads++
		}
	}
	if skipUploads != 2 {
		t.Errorf("GoReleaser manifest must route cask and formula uploads through the ship pipeline's signed-checksum flow (skip_upload: true twice), found %d", skipUploads)
	}
}

// releaseWorkflow reads the repository's release workflow.
func releaseWorkflow(t *testing.T) string {
	t.Helper()
	return readRepoFile(t, filepath.Join(".github", "workflows", "release.yml"))
}

// goreleaserManifest reads the repository's GoReleaser manifest.
func goreleaserManifest(t *testing.T) string {
	t.Helper()
	return readRepoFile(t, ".goreleaser.yaml")
}

// readRepoFile reads a repository file resolved relative to this package's
// directory, so the regression tests always pin the checked-in
// configuration regardless of the working directory the tests run from.
func readRepoFile(t *testing.T, relativePath string) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate release configuration regression test")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(testFile), "..", "..", relativePath))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(content)
}

// lineContaining returns the full line containing needle, or "" when no
// line contains it.
func lineContaining(content, needle string) string {
	index := strings.Index(content, needle)
	if index == -1 {
		return ""
	}
	start := strings.LastIndex(content[:index], "\n") + 1
	end := strings.Index(content[index:], "\n")
	if end == -1 {
		end = len(content)
	} else {
		end += index
	}
	return content[start:end]
}

// jobPermissionGrants collects the permission grants of the indented
// release-job permissions block.
func jobPermissionGrants(workflow string) []string {
	const blockHeader = "\n    permissions:\n"
	header := strings.Index(workflow, blockHeader)
	if header == -1 {
		return nil
	}
	var grants []string
	for _, line := range strings.Split(workflow[header+len(blockHeader):], "\n") {
		if !strings.HasPrefix(line, "      ") {
			break
		}
		grants = append(grants, strings.TrimSpace(line))
	}
	return grants
}

// goreleaserCustomBlock extracts the homebrew_casks custom_block literal,
// stopping at the first line indented no deeper than the block key.
func goreleaserCustomBlock(t *testing.T, manifest string) string {
	t.Helper()
	const blockKey = "custom_block: |"
	keyIndex := strings.Index(manifest, blockKey)
	if keyIndex == -1 {
		t.Fatal(".goreleaser.yaml is missing the cask custom_block")
	}
	lineStart := strings.LastIndex(manifest[:keyIndex], "\n") + 1
	keyLine := manifest[lineStart : keyIndex+len(blockKey)]
	keyIndent := keyLine[:len(keyLine)-len(strings.TrimLeft(keyLine, " "))]
	var block []string
	for _, line := range strings.Split(manifest[keyIndex+len(blockKey):], "\n")[1:] {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, keyIndent+" ") {
			break
		}
		block = append(block, line)
	}
	return strings.Join(block, "\n")
}

// unescapeGoReleaserTemplates resolves GoReleaser's backtick escaping so
// the custom_block can be compared with the transformer's literal output.
func unescapeGoReleaserTemplates(block string) string {
	for escaped, literal := range goreleaserEscapedTemplates {
		block = strings.ReplaceAll(block, escaped, literal)
	}
	return block
}

// dedentBlock strips the leading indentation of its first non-empty line
// from every line, normalizing embedded blocks to column zero.
func dedentBlock(block string) string {
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	indent := ""
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent = line[:len(line)-len(strings.TrimLeft(line, " "))]
		break
	}
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, indent)
	}
	return strings.Join(lines, "\n")
}
