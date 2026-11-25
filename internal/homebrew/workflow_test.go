package homebrew

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The release runs through jflowers/ship's go-release.yml, which patches the
// signed checksums into the cask and formula, audits and install-checks the
// cask, and pushes both (tested in ship, against this repo's cask and
// formula). What stays here is what this repo configures: the inputs it
// passes, and the cask's postflight_steps, which GoReleaser now writes.

// TestReleaseWorkflow_PublishesCaskAndFormulaThroughShip checks that
// release.yml hands ship the tap, the formula and an install check of the
// binary and the uf alias.
func TestReleaseWorkflow_PublishesCaskAndFormulaThroughShip(t *testing.T) {
	t.Parallel()

	var workflow struct {
		Jobs map[string]struct {
			Uses    string            `yaml:"uses"`
			With    map[string]string `yaml:"with"`
			Secrets map[string]string `yaml:"secrets"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(readRepoFile(t, ".github/workflows/release.yml"), &workflow); err != nil {
		t.Fatalf("parse release workflow: %v", err)
	}
	release, ok := workflow.Jobs["release"]
	if !ok {
		t.Fatal("release workflow has no release job")
	}
	if !strings.HasPrefix(release.Uses, "jflowers/ship/.github/workflows/go-release.yml@") {
		t.Fatalf("release job uses %q, want jflowers/ship's go-release.yml", release.Uses)
	}
	if sha := strings.TrimPrefix(release.Uses, "jflowers/ship/.github/workflows/go-release.yml@"); len(sha) != 40 {
		t.Errorf("go-release.yml must be pinned by full commit SHA, got %q", sha)
	}
	for input, want := range map[string]string{
		"tap":     "unbound-force/homebrew-tap",
		"formula": "unbound-force",
	} {
		if got := release.With[input]; got != want {
			t.Errorf("release job input %s = %q, want %q", input, got, want)
		}
	}
	check := release.With["cask_install_check"]
	for _, command := range []string{
		`test -x "$HOMEBREW_PREFIX/bin/unbound-force"`,
		`test -L "$HOMEBREW_PREFIX/bin/uf"`,
	} {
		if !strings.Contains(check, command) {
			t.Errorf("cask_install_check is missing %q", command)
		}
	}
	for _, secret := range []string{
		"MACOS_SIGN_P12", "MACOS_SIGN_PASSWORD", "MACOS_SIGN_IDENTITY",
		"MACOS_NOTARY_KEY", "MACOS_NOTARY_KEY_ID", "MACOS_NOTARY_ISSUER_ID",
		"HOMEBREW_TAP_TOKEN",
	} {
		if release.Secrets[secret] != "${{ secrets."+secret+" }}" {
			t.Errorf("release job must pass secret %s by name", secret)
		}
	}
}

// TestGoReleaser_CaskPostflightStepsMatchTransformer checks that the cask's
// custom_block, once GoReleaser renders it, is exactly the postflight_steps
// the transformer produced: quarantine removal and the uf symlink.
func TestGoReleaser_CaskPostflightStepsMatchTransformer(t *testing.T) {
	t.Parallel()

	var config struct {
		Casks []struct {
			Name        string         `yaml:"name"`
			CustomBlock string         `yaml:"custom_block"`
			Hooks       map[string]any `yaml:"hooks"`
		} `yaml:"homebrew_casks"`
	}
	if err := yaml.Unmarshal(readRepoFile(t, ".goreleaser.yaml"), &config); err != nil {
		t.Fatalf("parse .goreleaser.yaml: %v", err)
	}
	if len(config.Casks) != 1 || config.Casks[0].Name != "unbound-force" {
		t.Fatalf("want one homebrew_casks entry named unbound-force, got %+v", config.Casks)
	}
	cask := config.Casks[0]
	if len(cask.Hooks) != 0 {
		t.Error("the cask must not use hooks: GoReleaser renders them as the deprecated postflight block")
	}

	// GoReleaser templates the block: {{ `{{x}}` }} renders as {{x}}, and
	// each line is indented two spaces inside the cask.
	rendered := strings.NewReplacer(
		"{{ `{{staged_path}}` }}", "{{staged_path}}",
		"{{ `{{HOMEBREW_PREFIX}}` }}", "{{HOMEBREW_PREFIX}}",
	).Replace(strings.TrimRight(cask.CustomBlock, "\n"))
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = "  " + line
		}
	}
	if got := strings.Join(lines, "\n"); got != declarativePostflightSteps {
		t.Errorf("custom_block renders as\n%s\nwant the transformer's postflight_steps\n%s", got, declarativePostflightSteps)
	}
}

func readRepoFile(t *testing.T, path string) []byte {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(testFile), "..", "..", path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
}
