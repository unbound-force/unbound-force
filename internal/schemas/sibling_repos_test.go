package schemas_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
	"gopkg.in/yaml.v3"
)

func validSiblingDeclarationForTest() map[string]any {
	return map[string]any{
		"version": 1,
		"siblings": []any{
			map[string]any{
				"name":           "gaze",
				"url":            "https://github.com/unbound-force/gaze.git",
				"local_paths":    []any{".uf/sibling-cache/gaze"},
				"clone_path":     ".uf/sibling-clones/gaze",
				"cache_path":     ".uf/sibling-cache/gaze",
				"default_branch": "main",
				"role":           "sibling",
				"fetch":          "clone-on-review",
				"contracts":      []any{"README.md", "docs/*.md"},
				"notes":          "Quality contracts.",
			},
		},
	}
}

func validateSiblingDeclarationForTest(t *testing.T, declaration map[string]any) error {
	t.Helper()
	schemaData, err := os.ReadFile(filepath.Join(repoSchemasDir(), "sibling-repos", "v1.0.0.schema.json"))
	if err != nil {
		t.Fatalf("read sibling schema: %v", err)
	}
	data, err := json.Marshal(declaration)
	if err != nil {
		t.Fatalf("marshal sibling declaration: %v", err)
	}
	return schemas.ValidateBytes(schemaData, data)
}

func siblingEntryForTest(t *testing.T, declaration map[string]any) map[string]any {
	t.Helper()
	siblings, ok := declaration["siblings"].([]any)
	if !ok || len(siblings) != 1 {
		t.Fatalf("test declaration siblings = %#v, want one entry", declaration["siblings"])
	}
	entry, ok := siblings[0].(map[string]any)
	if !ok {
		t.Fatalf("test sibling entry type = %T, want map", siblings[0])
	}
	return entry
}

func TestSiblingReposSchema_Boundaries(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "wrong version", mutate: func(root map[string]any) { root["version"] = 2 }},
		{name: "missing siblings", mutate: func(root map[string]any) { delete(root, "siblings") }},
		{name: "unknown root field", mutate: func(root map[string]any) { root["unknown"] = true }},
		{name: "too many siblings", mutate: func(root map[string]any) {
			entry := siblingEntryForTest(t, root)
			items := make([]any, 33)
			for index := range items {
				clone := make(map[string]any, len(entry))
				for key, value := range entry {
					clone[key] = value
				}
				clone["name"] = "repo-" + string(rune('a'+index%26)) + strings.Repeat("x", index/26)
				items[index] = clone
			}
			root["siblings"] = items
		}},
		{name: "missing required entry field", mutate: func(root map[string]any) { delete(siblingEntryForTest(t, root), "contracts") }},
		{name: "unknown entry field", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["unknown"] = true }},
		{name: "invalid name", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["name"] = "Gaze" }},
		{name: "empty name", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["name"] = "" }},
		{name: "long name", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["name"] = "a" + strings.Repeat("b", 63) }},
		{name: "null name", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["name"] = nil }},
		{name: "empty URL", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["url"] = "" }},
		{name: "long URL", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["url"] = "https://github.com/unbound-force/" + strings.Repeat("a", 480)
		}},
		{name: "HTTP URL", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["url"] = "http://github.com/unbound-force/gaze"
		}},
		{name: "credentialed URL", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["url"] = "https://token@github.com/unbound-force/gaze"
		}},
		{name: "non GitHub URL", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["url"] = "https://example.com/unbound-force/gaze"
		}},
		{name: "URL query", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["url"] = "https://github.com/unbound-force/gaze?token=x"
		}},
		{name: "empty branch", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["default_branch"] = "" }},
		{name: "non ASCII branch", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["default_branch"] = "maín" }},
		{name: "long branch", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["default_branch"] = strings.Repeat("a", 129) }},
		{name: "invalid role", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["role"] = "consumer" }},
		{name: "invalid fetch", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["fetch"] = "always" }},
		{name: "empty contracts", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["contracts"] = []any{} }},
		{name: "too many contracts", mutate: func(root map[string]any) {
			items := make([]any, 21)
			for index := range items {
				items[index] = "docs/file-" + string(rune('a'+index)) + ".md"
			}
			siblingEntryForTest(t, root)["contracts"] = items
		}},
		{name: "duplicate contracts", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["contracts"] = []any{"README.md", "README.md"} }},
		{name: "absolute contract", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["contracts"] = []any{"/README.md"} }},
		{name: "contract parent escape", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["contracts"] = []any{"../README.md"} }},
		{name: "contract empty segment", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["contracts"] = []any{"docs//README.md"} }},
		{name: "long contract", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["contracts"] = []any{strings.Repeat("a", 257)} }},
		{name: "too many local paths", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["local_paths"] = []any{"a", "b", "c", "d", "e"}
		}},
		{name: "duplicate local paths", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["local_paths"] = []any{"cache", "cache"} }},
		{name: "empty local path", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["local_paths"] = []any{""} }},
		{name: "local path escape", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["local_paths"] = []any{"../gaze"} }},
		{name: "long local path", mutate: func(root map[string]any) {
			siblingEntryForTest(t, root)["local_paths"] = []any{strings.Repeat("a", 513)}
		}},
		{name: "clone path absolute", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["clone_path"] = "/tmp/gaze" }},
		{name: "clone path escape", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["clone_path"] = "../gaze" }},
		{name: "long clone path", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["clone_path"] = strings.Repeat("a", 513) }},
		{name: "cache path dot", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["cache_path"] = "." }},
		{name: "long cache path", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["cache_path"] = strings.Repeat("a", 513) }},
		{name: "empty notes", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["notes"] = "" }},
		{name: "long notes", mutate: func(root map[string]any) { siblingEntryForTest(t, root)["notes"] = strings.Repeat("a", 1025) }},
	}

	if err := validateSiblingDeclarationForTest(t, validSiblingDeclarationForTest()); err != nil {
		t.Fatalf("valid declaration failed: %v", err)
	}
	if err := validateSiblingDeclarationForTest(t, map[string]any{"version": 1, "siblings": []any{}}); err != nil {
		t.Fatalf("empty declaration failed: %v", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			declaration := validSiblingDeclarationForTest()
			test.mutate(declaration)
			if err := validateSiblingDeclarationForTest(t, declaration); err == nil {
				t.Fatal("expected schema validation failure")
			}
		})
	}
}

func TestSiblingReposSchema_EveryEntryFieldIsRequired(t *testing.T) {
	for _, field := range []string{"name", "url", "default_branch", "role", "fetch", "contracts"} {
		t.Run(field, func(t *testing.T) {
			declaration := validSiblingDeclarationForTest()
			delete(siblingEntryForTest(t, declaration), field)
			if err := validateSiblingDeclarationForTest(t, declaration); err == nil {
				t.Fatalf("expected missing %s to fail schema validation", field)
			}
		})
	}
}

func TestSiblingReposSchema_LiveAndScaffoldDeclarationsValidate(t *testing.T) {
	root := filepath.Join(repoSchemasDir(), "..")
	paths := []string{
		filepath.Join(root, ".uf", "sibling-repos.yaml"),
		filepath.Join(root, "internal", "scaffold", "assets", "uf", "sibling-repos.yaml"),
	}
	for _, path := range paths {
		t.Run(filepath.Base(filepath.Dir(path))+"/"+filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read declaration: %v", err)
			}
			var declaration map[string]any
			if err := yaml.Unmarshal(data, &declaration); err != nil {
				t.Fatalf("parse declaration YAML: %v", err)
			}
			if err := validateSiblingDeclarationForTest(t, declaration); err != nil {
				t.Fatalf("validate declaration: %v", err)
			}
		})
	}
}

func TestSiblingReposScaffold_HasNoActiveEntries(t *testing.T) {
	path := filepath.Join(repoSchemasDir(), "..", "internal", "scaffold", "assets", "uf", "sibling-repos.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read sibling scaffold: %v", err)
	}
	var declaration struct {
		Version  int              `yaml:"version"`
		Siblings []map[string]any `yaml:"siblings"`
	}
	if err := yaml.Unmarshal(data, &declaration); err != nil {
		t.Fatalf("parse sibling scaffold: %v", err)
	}
	if declaration.Version != 1 || len(declaration.Siblings) != 0 {
		t.Errorf("scaffold declaration = version %d, %d siblings; want version 1 and zero siblings", declaration.Version, len(declaration.Siblings))
	}
}
