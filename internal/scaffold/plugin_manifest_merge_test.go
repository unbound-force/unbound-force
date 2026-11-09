package scaffold

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pluginManifestFields is the subset of .opencode/package.json the merge
// tests inspect.
type pluginManifestFields struct {
	Name            string            `json:"name"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Overrides       map[string]string `json:"overrides"`
}

func decodePluginManifest(t *testing.T, data []byte) pluginManifestFields {
	t.Helper()
	var m pluginManifestFields
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("merged manifest is not valid JSON: %v\n%s", err, data)
	}
	return m
}

func loadShippedPluginManifest(t *testing.T) ([]byte, pluginManifestFields) {
	t.Helper()
	shipped, err := assetContent(pluginManifestAsset)
	if err != nil {
		t.Fatalf("read shipped manifest: %v", err)
	}
	return shipped, decodePluginManifest(t, shipped)
}

func TestPreservePluginRuntimeDependencies_Cases(t *testing.T) {
	shipped, pins := loadShippedPluginManifest(t)

	tests := []struct {
		name      string
		existing  string
		unchanged bool              // want the shipped bytes back
		want      map[string]string // merged runtime dependencies otherwise
	}{
		{name: "unparsable target", existing: "not json", unchanged: true},
		{name: "non-string dependency value", existing: `{"dependencies": {"left-pad": 1}}`, unchanged: true},
		{name: "no dependencies", existing: `{"name": "x"}`, unchanged: true},
		{name: "empty dependencies", existing: `{"dependencies": {}}`, unchanged: true},
		{name: "only shipped pins", existing: `{"dependencies": {"zod": "4.6.6", "@opencode-ai/plugin": "1.19.0"}}`, unchanged: true},
		{name: "drifted shipped pin only", existing: `{"dependencies": {"zod": "4.2.0"}}`, unchanged: true},
		{
			name:     "user package kept and shipped pins restored",
			existing: `{"dependencies": {"left-pad": "1.3.0"}}`,
			want:     map[string]string{"@opencode-ai/plugin": pins.Dependencies["@opencode-ai/plugin"], "left-pad": "1.3.0", "zod": pins.Dependencies["zod"]},
		},
		{
			name:     "drifted pin bumped beside user package",
			existing: `{"dependencies": {"zod": "4.2.0", "@opencode-ai/plugin": "0.1.0", "my-plugin": "^2.0.0"}}`,
			want:     map[string]string{"@opencode-ai/plugin": pins.Dependencies["@opencode-ai/plugin"], "my-plugin": "^2.0.0", "zod": pins.Dependencies["zod"]},
		},
		{
			name:     "range characters stay literal",
			existing: `{"dependencies": {"a-tool": ">=1.0.0 <2"}}`,
			want:     map[string]string{"@opencode-ai/plugin": pins.Dependencies["@opencode-ai/plugin"], "a-tool": ">=1.0.0 <2", "zod": pins.Dependencies["zod"]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preservePluginRuntimeDependencies(shipped, []byte(tt.existing))
			if tt.unchanged {
				if !bytes.Equal(got, shipped) {
					t.Errorf("got %s, want the shipped manifest unchanged", got)
				}
				return
			}
			m := decodePluginManifest(t, got)
			if len(m.Dependencies) != len(tt.want) {
				t.Errorf("dependencies = %v, want %v", m.Dependencies, tt.want)
			}
			for name, version := range tt.want {
				if m.Dependencies[name] != version {
					t.Errorf("dependencies[%q] = %q, want %q", name, m.Dependencies[name], version)
				}
			}
			// Everything outside the dependencies block stays tool-owned.
			if m.Name != pins.Name || len(m.Overrides) != len(pins.Overrides) {
				t.Errorf("tool-owned fields changed: %+v", m)
			}
			for name, version := range pins.Overrides {
				if m.Overrides[name] != version {
					t.Errorf("overrides[%q] = %q, want %q", name, m.Overrides[name], version)
				}
			}
			assertSortedDependencyKeys(t, got)
			if bytes.Contains(got, []byte("\\u00")) {
				t.Errorf("version ranges must not be HTML-escaped:\n%s", got)
			}
			// Deterministic: a second merge over its own output is a no-op.
			if again := preservePluginRuntimeDependencies(shipped, got); !bytes.Equal(again, got) {
				t.Errorf("merge is not idempotent:\nfirst:  %s\nsecond: %s", got, again)
			}
		})
	}
}

func TestPreservePluginRuntimeDependencies_AnchorMiss(t *testing.T) {
	existing := []byte(`{"dependencies": {"left-pad": "1.3.0"}}`)
	for _, shipped := range []string{
		"{}\n",
		"not json",
		`{"dependencies": {"zod": "4.6.6"}}`,
		"{\n  \"dependencies\": {\n    \"zod\": \"4.6.6\"",
	} {
		if got := preservePluginRuntimeDependencies([]byte(shipped), existing); string(got) != shipped {
			t.Errorf("shipped %q: got %s, want it returned unchanged", shipped, got)
		}
	}
}

func TestMergePluginManifest_RestoresRepositoryManifest(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}
	repo, err := os.ReadFile(filepath.Join(root, ".opencode", "package.json"))
	if err != nil {
		t.Fatalf("read repository manifest: %v", err)
	}
	shipped, _ := loadShippedPluginManifest(t)
	if got := mergePluginManifest(shipped, repo); !bytes.Equal(got, repo) {
		t.Errorf("merged manifest differs from the repository manifest:\n%s", got)
	}
}

func TestMergePluginManifest_KeepsRuntimeAndDevDependencies(t *testing.T) {
	shipped, pins := loadShippedPluginManifest(t)
	existing := `{"dependencies": {"zod": "4.2.0", "left-pad": "1.3.0"},
	  "devDependencies": {"vitest": "5.0.3"},
	  "overrides": {"stale-override": "1.0.0"}}`
	got := mergePluginManifest(shipped, []byte(existing))
	m := decodePluginManifest(t, got)
	if m.Dependencies["left-pad"] != "1.3.0" || m.Dependencies["zod"] != pins.Dependencies["zod"] {
		t.Errorf("dependencies = %v", m.Dependencies)
	}
	if m.DevDependencies["vitest"] != "5.0.3" || len(m.DevDependencies) != 1 {
		t.Errorf("devDependencies = %v", m.DevDependencies)
	}
	if _, ok := m.Overrides["stale-override"]; ok || len(m.Overrides) != len(pins.Overrides) {
		t.Errorf("overrides must stay tool-owned, got %v", m.Overrides)
	}
}

// assertSortedDependencyKeys checks that the dependencies block lists its
// keys in ascending byte order.
func assertSortedDependencyKeys(t *testing.T, manifest []byte) {
	t.Helper()
	s := string(manifest)
	start := strings.Index(s, "\"dependencies\": {\n")
	if start < 0 {
		t.Fatalf("dependencies block not found:\n%s", manifest)
	}
	end := strings.Index(s[start:], "\n  }")
	if end < 0 {
		t.Fatalf("dependencies block not closed:\n%s", manifest)
	}
	var prev string
	for _, line := range strings.Split(s[start:start+end], "\n")[1:] {
		key := strings.SplitN(strings.TrimSpace(line), ":", 2)[0]
		if prev != "" && key <= prev {
			t.Errorf("dependency keys not sorted: %s after %s", key, prev)
		}
		prev = key
	}
}

// TestSCFR003_UserRuntimeDependencySurvivesInit is the acceptance test for
// the merge-plugin-runtime-deps scenarios "User-added runtime dependency
// survives uf init", "Drifted shipped pin is restored", and "Unchanged
// merged manifest is skipped", driven through Run().
func TestSCFR003_UserRuntimeDependencySurvivesInit(t *testing.T) {
	const manifestRel = ".opencode/package.json"
	shipped, pins := loadShippedPluginManifest(t)

	writeAndRun := func(t *testing.T, existing string, force bool) (*Result, []byte) {
		t.Helper()
		dir := t.TempDir()
		manifestPath := filepath.Join(dir, ".opencode", "package.json")
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
			t.Fatalf("create .opencode: %v", err)
		}
		if err := os.WriteFile(manifestPath, []byte(existing), 0o644); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
		var buf bytes.Buffer
		result, err := runScaffoldWithoutExternalTools(t, Options{
			TargetDir: dir,
			Version:   "1.0.0",
			Force:     force,
			Stdout:    &buf,
		})
		if err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		got, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("read manifest: %v", err)
		}
		return result, got
	}
	has := func(list []string, want string) bool {
		for _, item := range list {
			if item == want {
				return true
			}
		}
		return false
	}

	t.Run("user-added runtime dependency survives uf init", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			_, got := writeAndRun(t, `{"dependencies": {"left-pad": "1.3.0"}}`, force)
			m := decodePluginManifest(t, got)
			if m.Dependencies["left-pad"] != "1.3.0" {
				t.Errorf("force=%v: left-pad = %q, want 1.3.0 kept", force, m.Dependencies["left-pad"])
			}
			for name, version := range pins.Dependencies {
				if m.Dependencies[name] != version {
					t.Errorf("force=%v: %s = %q, want shipped pin %q", force, name, m.Dependencies[name], version)
				}
			}
		}
	})

	t.Run("drifted shipped pin is restored", func(t *testing.T) {
		result, got := writeAndRun(t, `{"dependencies": {"zod": "4.2.0", "left-pad": "1.3.0"}}`, false)
		m := decodePluginManifest(t, got)
		if m.Dependencies["zod"] != pins.Dependencies["zod"] || m.Dependencies["@opencode-ai/plugin"] != pins.Dependencies["@opencode-ai/plugin"] {
			t.Errorf("shipped pins not restored: %v", m.Dependencies)
		}
		if !has(result.Updated, manifestRel) {
			t.Errorf("%s not reported as updated: %+v", manifestRel, result.Updated)
		}
	})

	t.Run("unchanged merged manifest is skipped", func(t *testing.T) {
		merged := preservePluginRuntimeDependencies(shipped, []byte(`{"dependencies": {"left-pad": "1.3.0"}}`))
		result, got := writeAndRun(t, string(merged), false)
		if !bytes.Equal(got, merged) {
			t.Errorf("merged manifest rewritten:\n%s", got)
		}
		if !has(result.Skipped, manifestRel) {
			t.Errorf("%s not reported as skipped: %+v", manifestRel, result.Skipped)
		}
	})

	t.Run("unparsable manifest is restored", func(t *testing.T) {
		result, got := writeAndRun(t, "not json", false)
		if !bytes.Equal(got, shipped) {
			t.Errorf("unparsable manifest not restored to shipped:\n%s", got)
		}
		if !has(result.Updated, manifestRel) {
			t.Errorf("%s not reported as updated: %+v", manifestRel, result.Updated)
		}
	})
}
