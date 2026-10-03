package schemas_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

type reviewerManifestFixture struct {
	Version   int                    `yaml:"version"`
	Reviewers []reviewerEntryFixture `yaml:"reviewers"`
}

type reviewerEntryFixture struct {
	Agent      string   `yaml:"agent"`
	Capability string   `yaml:"capability"`
	Scopes     []string `yaml:"scopes"`
}

func knownReviewerEntriesForTest() []reviewerEntryFixture {
	return []reviewerEntryFixture{
		{Agent: "divisor-adversary", Capability: "review", Scopes: []string{"security", "dependencies", "standard"}},
		{Agent: "divisor-architect", Capability: "review", Scopes: []string{"standard", "cli-ux", "ci-cd", "documentation"}},
		{Agent: "divisor-curator", Capability: "review", Scopes: []string{"documentation"}},
		{Agent: "divisor-guard", Capability: "review", Scopes: []string{"standard", "cli-ux", "documentation"}},
		{Agent: "divisor-sre", Capability: "review", Scopes: []string{"ci-cd", "dependencies", "security"}},
		{Agent: "divisor-testing", Capability: "review", Scopes: []string{"test-quality"}},
		{Agent: "divisor-envoy", Capability: "content", Scopes: []string{}},
		{Agent: "divisor-herald", Capability: "content", Scopes: []string{}},
		{Agent: "divisor-scribe", Capability: "content", Scopes: []string{}},
	}
}

func parseReviewerManifestForTest(schemaData, manifestData []byte) (reviewerManifestFixture, error) {
	parsed, err := parser.ParseBytes(manifestData, 0)
	if err != nil {
		return reviewerManifestFixture{}, fmt.Errorf("parse reviewer manifest YAML: %w", err)
	}
	for _, nodeType := range []ast.NodeType{ast.AnchorType, ast.AliasType, ast.MergeKeyType} {
		if len(ast.FilterFile(nodeType, parsed)) != 0 {
			return reviewerManifestFixture{}, fmt.Errorf("reviewer manifest contains forbidden YAML %s", nodeType)
		}
	}

	manifestJSON, err := yaml.YAMLToJSON(manifestData)
	if err != nil {
		return reviewerManifestFixture{}, fmt.Errorf("convert reviewer manifest to JSON: %w", err)
	}
	if err := schemas.ValidateBytes(schemaData, manifestJSON); err != nil {
		return reviewerManifestFixture{}, fmt.Errorf("validate reviewer manifest schema: %w", err)
	}

	var manifest reviewerManifestFixture
	if err := yaml.UnmarshalWithOptions(manifestData, &manifest, yaml.Strict()); err != nil {
		return reviewerManifestFixture{}, fmt.Errorf("decode reviewer manifest: %w", err)
	}
	if err := validateReviewerManifestSemanticsForTest(manifest); err != nil {
		return reviewerManifestFixture{}, err
	}
	return manifest, nil
}

func validateReviewerManifestSemanticsForTest(manifest reviewerManifestFixture) error {
	seen := make(map[string]reviewerEntryFixture, len(manifest.Reviewers))
	for _, entry := range manifest.Reviewers {
		if previous, exists := seen[entry.Agent]; exists {
			if reflect.DeepEqual(previous, entry) {
				return fmt.Errorf("duplicate reviewer entry for %s", entry.Agent)
			}
			return fmt.Errorf("conflicting reviewer entries for %s", entry.Agent)
		}
		seen[entry.Agent] = entry
	}

	for _, expected := range knownReviewerEntriesForTest() {
		actual, exists := seen[expected.Agent]
		if !exists {
			return fmt.Errorf("missing known reviewer entry for %s", expected.Agent)
		}
		if !reflect.DeepEqual(actual, expected) {
			return fmt.Errorf("known reviewer entry for %s conflicts with canonical policy", expected.Agent)
		}
	}
	return nil
}

func eligibleReviewersForTest(
	manifest reviewerManifestFixture,
	discovered []string,
	categories []string,
	userFacing bool,
) ([]string, error) {
	if err := validateReviewerManifestSemanticsForTest(manifest); err != nil {
		return nil, err
	}

	entries := make(map[string]reviewerEntryFixture, len(manifest.Reviewers))
	for _, entry := range manifest.Reviewers {
		entries[entry.Agent] = entry
	}
	categorySet := make(map[string]bool, len(categories))
	for _, category := range categories {
		categorySet[category] = true
	}

	ordered := append([]string(nil), discovered...)
	sort.Strings(ordered)
	eligible := make([]string, 0, len(ordered))
	for _, agent := range ordered {
		entry, exists := entries[agent]
		if !exists {
			return nil, fmt.Errorf("discovered agent %s has no manifest entry", agent)
		}
		if entry.Capability == "content" {
			continue
		}
		if agent == "divisor-adversary" || agent == "divisor-guard" {
			eligible = append(eligible, agent)
			continue
		}
		if agent == "divisor-curator" {
			if categorySet["documentation"] || userFacing {
				eligible = append(eligible, agent)
			}
			continue
		}
		for _, scope := range entry.Scopes {
			if categorySet[scope] {
				eligible = append(eligible, agent)
				break
			}
		}
	}
	return eligible, nil
}

func loadCanonicalReviewerManifestForTest(t *testing.T) reviewerManifestFixture {
	t.Helper()
	schemasDir := repoSchemasDir()
	schemaData, err := os.ReadFile(filepath.Join(schemasDir, "reviewer-capabilities", "v1.0.0.schema.json"))
	if err != nil {
		t.Fatalf("read reviewer-capabilities schema: %v", err)
	}
	manifestData, err := os.ReadFile(filepath.Join(schemasDir, "..", ".uf", "reviewer-capabilities.yaml"))
	if err != nil {
		t.Fatalf("read canonical reviewer manifest: %v", err)
	}
	manifest, err := parseReviewerManifestForTest(schemaData, manifestData)
	if err != nil {
		t.Fatalf("parse canonical reviewer manifest: %v", err)
	}
	return manifest
}

func TestReviewerCapabilities_CanonicalManifestHasExactOrderAndScopes(t *testing.T) {
	manifest := loadCanonicalReviewerManifestForTest(t)
	want := knownReviewerEntriesForTest()
	if !reflect.DeepEqual(manifest.Reviewers, want) {
		t.Errorf("canonical reviewers = %#v, want %#v", manifest.Reviewers, want)
	}

	reordered := manifest
	reordered.Reviewers = append([]reviewerEntryFixture(nil), manifest.Reviewers...)
	reordered.Reviewers[0].Scopes = []string{"dependencies", "security", "standard"}
	if err := validateReviewerManifestSemanticsForTest(reordered); err == nil {
		t.Error("expected reordered known scopes to conflict with canonical policy")
	}
}

func TestReviewerCapabilities_EligibilityPolicy(t *testing.T) {
	manifest := loadCanonicalReviewerManifestForTest(t)
	tests := []struct {
		name       string
		discovered []string
		categories []string
		userFacing bool
		want       []string
	}{
		{
			name:       "adversary and guard are always eligible",
			discovered: []string{"divisor-testing", "divisor-guard", "divisor-adversary"},
			categories: []string{"test-quality"},
			want:       []string{"divisor-adversary", "divisor-guard", "divisor-testing"},
		},
		{
			name: "category intersections use declared scopes",
			discovered: []string{
				"divisor-testing", "divisor-sre", "divisor-guard",
				"divisor-curator", "divisor-architect", "divisor-adversary",
			},
			categories: []string{"ci-cd"},
			want: []string{
				"divisor-adversary", "divisor-architect", "divisor-guard", "divisor-sre",
			},
		},
		{
			name:       "content personas are excluded",
			discovered: []string{"divisor-scribe", "divisor-guard", "divisor-herald", "divisor-envoy"},
			categories: []string{"documentation"},
			want:       []string{"divisor-guard"},
		},
		{
			name:       "curator is pruned outside documentation",
			discovered: []string{"divisor-curator", "divisor-guard"},
			categories: []string{"standard"},
			want:       []string{"divisor-guard"},
		},
		{
			name:       "curator intersects documentation",
			discovered: []string{"divisor-guard", "divisor-curator"},
			categories: []string{"documentation"},
			want:       []string{"divisor-curator", "divisor-guard"},
		},
		{
			name:       "curator accepts user-facing classification",
			discovered: []string{"divisor-guard", "divisor-curator"},
			categories: []string{"standard"},
			userFacing: true,
			want:       []string{"divisor-curator", "divisor-guard"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first, err := eligibleReviewersForTest(manifest, test.discovered, test.categories, test.userFacing)
			if err != nil {
				t.Fatalf("select eligible reviewers: %v", err)
			}
			second, err := eligibleReviewersForTest(manifest, test.discovered, test.categories, test.userFacing)
			if err != nil {
				t.Fatalf("repeat eligible reviewer selection: %v", err)
			}
			if !reflect.DeepEqual(first, test.want) {
				t.Errorf("eligible reviewers = %v, want %v", first, test.want)
			}
			if !reflect.DeepEqual(second, first) {
				t.Errorf("repeat eligible reviewers = %v, want deterministic %v", second, first)
			}
		})
	}
}

func TestReviewerCapabilities_UnknownAgentsRequireExplicitEntry(t *testing.T) {
	manifest := loadCanonicalReviewerManifestForTest(t)
	if _, err := eligibleReviewersForTest(
		manifest,
		[]string{"divisor-performance"},
		[]string{"standard"},
		false,
	); err == nil {
		t.Fatal("expected an unmanifested unknown reviewer to fail selection")
	}

	manifest.Reviewers = append(manifest.Reviewers, reviewerEntryFixture{
		Agent: "divisor-performance", Capability: "review", Scopes: []string{"standard", "ci-cd"},
	})
	got, err := eligibleReviewersForTest(
		manifest,
		[]string{"divisor-performance", "divisor-guard"},
		[]string{"standard"},
		false,
	)
	if err != nil {
		t.Fatalf("select explicitly manifested unknown reviewer: %v", err)
	}
	want := []string{"divisor-guard", "divisor-performance"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("eligible unknown reviewers = %v, want %v", got, want)
	}
}

func TestReviewerCapabilities_InvalidPolicyFixturesFailClosed(t *testing.T) {
	schemasDir := repoSchemasDir()
	schemaData, err := os.ReadFile(filepath.Join(schemasDir, "reviewer-capabilities", "v1.0.0.schema.json"))
	if err != nil {
		t.Fatalf("read reviewer-capabilities schema: %v", err)
	}

	fixtures := []struct {
		name        string
		wantMessage string
	}{
		{name: "invalid-alias.yaml", wantMessage: "forbidden YAML"},
		{name: "invalid-conflicting-agent.yaml", wantMessage: "conflicting reviewer entries"},
		{name: "invalid-duplicate-agent.yaml", wantMessage: "duplicate reviewer entry"},
		{name: "invalid-malformed.yaml", wantMessage: "parse reviewer manifest YAML"},
		{name: "invalid-missing-known-agent.yaml", wantMessage: "missing known reviewer entry"},
	}
	fixturesDir := filepath.Join(schemasDir, "reviewer-capabilities", "policy-fixtures")
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(fixturesDir, fixture.name))
			if err != nil {
				t.Fatalf("read policy fixture: %v", err)
			}
			_, err = parseReviewerManifestForTest(schemaData, data)
			if err == nil {
				t.Fatal("expected invalid policy fixture to fail closed")
			}
			if !strings.Contains(err.Error(), fixture.wantMessage) {
				t.Errorf("fixture error = %q, want substring %q", err, fixture.wantMessage)
			}
		})
	}
}

func TestReviewerCapabilities_AgentFrontmatterContainsNoPolicyMetadata(t *testing.T) {
	root := filepath.Join(repoSchemasDir(), "..")
	for _, entry := range knownReviewerEntriesForTest() {
		t.Run(entry.Agent, func(t *testing.T) {
			path := filepath.Join(root, ".opencode", "agents", entry.Agent+".md")
			frontmatter := readAgentFrontmatterForTest(t, path)
			for _, forbidden := range []string{"capability", "reviewer_capability", "reviewer_scopes", "scopes"} {
				if _, exists := frontmatter[forbidden]; exists {
					t.Errorf("agent frontmatter contains reviewer policy key %q", forbidden)
				}
			}
		})
	}
}

func TestReviewerCapabilities_OpenCodeLoadKeepsProviderOptionsEmpty(t *testing.T) {
	fixtureRoot := filepath.Join(repoSchemasDir(), "reviewer-capabilities", "fixtures", "opencode-load")
	agentSource := filepath.Join(fixtureRoot, ".opencode", "agents", "divisor-adversary.md")
	frontmatter := readAgentFrontmatterForTest(t, agentSource)
	for _, forbidden := range []string{"capability", "reviewer_capability", "reviewer_scopes", "scopes"} {
		if _, exists := frontmatter[forbidden]; exists {
			t.Fatalf("load fixture frontmatter contains reviewer policy key %q", forbidden)
		}
	}

	opencodePath, err := exec.LookPath("opencode")
	if err != nil {
		t.Skip("opencode is not installed; static frontmatter boundary validated")
	}

	scratch := t.TempDir()
	copyReviewerLoadFixtureForTest(t, fixtureRoot, scratch)
	isolatedHome := filepath.Join(scratch, "home")
	if err := os.MkdirAll(isolatedHome, 0o755); err != nil {
		t.Fatalf("create isolated OpenCode home: %v", err)
	}

	command := exec.Command(opencodePath, "debug", "agent", "divisor-adversary", "--pure")
	command.Dir = scratch
	command.Env = isolatedOpenCodeEnvForTest(isolatedHome)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("provider-free OpenCode agent load failed: %v\n%s", err, output)
	}

	var loaded map[string]interface{}
	if err := json.Unmarshal(output, &loaded); err != nil {
		t.Fatalf("parse OpenCode agent output: %v\n%s", err, output)
	}
	if got := loaded["name"]; got != "divisor-adversary" {
		t.Errorf("loaded agent name = %v, want divisor-adversary", got)
	}
	options, ok := loaded["options"].(map[string]interface{})
	if !ok {
		t.Fatalf("loaded provider options type = %T, want object", loaded["options"])
	}
	if len(options) != 0 {
		t.Errorf("loaded provider options = %v, want empty object", options)
	}
	for _, forbidden := range []string{"capability", "reviewer_capability", "reviewer_scopes", "scopes"} {
		if _, exists := loaded[forbidden]; exists {
			t.Errorf("OpenCode agent output contains reviewer policy key %q", forbidden)
		}
	}
}

func readAgentFrontmatterForTest(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read agent fixture %s: %v", path, err)
	}
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) != 3 || strings.TrimSpace(parts[0]) != "" {
		t.Fatalf("agent %s has malformed frontmatter delimiters", path)
	}
	var frontmatter map[string]interface{}
	if err := yaml.UnmarshalWithOptions([]byte(parts[1]), &frontmatter, yaml.Strict()); err != nil {
		t.Fatalf("parse agent frontmatter %s: %v", path, err)
	}
	return frontmatter
}

func copyReviewerLoadFixtureForTest(t *testing.T, sourceRoot, targetRoot string) {
	t.Helper()
	paths := []string{
		filepath.Join(".opencode", "agents", "divisor-adversary.md"),
		filepath.Join(".uf", "reviewer-capabilities.yaml"),
	}
	for _, relativePath := range paths {
		data, err := os.ReadFile(filepath.Join(sourceRoot, relativePath))
		if err != nil {
			t.Fatalf("read OpenCode load fixture %s: %v", relativePath, err)
		}
		targetPath := filepath.Join(targetRoot, relativePath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			t.Fatalf("create OpenCode fixture directory: %v", err)
		}
		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			t.Fatalf("write OpenCode load fixture %s: %v", relativePath, err)
		}
	}
}

func isolatedOpenCodeEnvForTest(home string) []string {
	prefixes := []string{"HOME=", "XDG_CONFIG_HOME=", "XDG_DATA_HOME=", "XDG_CACHE_HOME=", "XDG_STATE_HOME="}
	environment := make([]string, 0, len(os.Environ())+5)
	for _, value := range os.Environ() {
		excluded := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(value, prefix) {
				excluded = true
				break
			}
		}
		if !excluded {
			environment = append(environment, value)
		}
	}
	return append(environment,
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, "config"),
		"XDG_DATA_HOME="+filepath.Join(home, "data"),
		"XDG_CACHE_HOME="+filepath.Join(home, "cache"),
		"XDG_STATE_HOME="+filepath.Join(home, "state"),
	)
}
