package coveragegate

import (
	"strings"
	"testing"
)

const validManifest = `{
  "version": 1,
  "scopes": {
    "helper": {"threshold": 80, "packages": ["internal/scaffold", "internal/doctor", "internal/artifacts", "internal/schemas"]},
    "artifact": {"threshold": 90, "packages": ["internal/artifactpkg"]},
    "schema": {"threshold": 90, "packages": ["internal/schemapkg"]}
  }
}`

func TestParseManifest_Valid(t *testing.T) {
	manifest, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("ParseManifest returned error for valid input: %v", err)
	}
	if manifest.Version != 1 {
		t.Fatalf("version = %d, want 1", manifest.Version)
	}
	if got := manifest.Scopes["helper"].Threshold; got != 80 {
		t.Fatalf("helper threshold = %v, want 80", got)
	}
	if got := manifest.Scopes["artifact"].Threshold; got != 90 {
		t.Fatalf("artifact threshold = %v, want 90", got)
	}
}

func TestParseManifest_RejectsInvalid(t *testing.T) {
	tests := []struct {
		name    string
		doc     string
		wantSub string
	}{
		{"unsupported version", `{"version": 2, "scopes": {}}`, "unsupported"},
		{"no scopes", `{"version": 1, "scopes": {}}`, "no scopes"},
		{"non-numeric threshold", `{"version": 1, "scopes": {"helper": {"threshold": "eighty", "packages": []}}}`, "decode"},
		{"threshold out of range", `{"version": 1, "scopes": {"helper": {"threshold": 101, "packages": []}}}`, "outside 0-100"},
		{"empty scope name", `{"version": 1, "scopes": {"": {"threshold": 80, "packages": []}}}`, "empty name"},
		{"empty package prefix", `{"version": 1, "scopes": {"helper": {"threshold": 80, "packages": [""]}}}`, "empty package"},
		{
			"multiply matched package",
			`{"version": 1, "scopes": {
				"helper": {"threshold": 80, "packages": ["internal/scaffold"]},
				"schema": {"threshold": 90, "packages": ["internal/scaffold"]}
			}}`,
			"multiply matched",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(test.doc))
			if err == nil {
				t.Fatalf("ParseManifest(%s) = nil error, want error containing %q", test.name, test.wantSub)
			}
			if !strings.Contains(err.Error(), test.wantSub) {
				t.Fatalf("ParseManifest(%s) error %q does not contain %q", test.name, err, test.wantSub)
			}
		})
	}
}

func TestManifestClassify(t *testing.T) {
	manifest, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	tests := []struct {
		pkg       string
		wantScope string
	}{
		{"internal/scaffold", "helper"},
		{"internal/scaffold/assets", "helper"},
		{"internal/artifacts", "helper"},
		{"internal/schemas", "helper"},
		{"internal/artifactpkg", "artifact"},
		{"internal/schemapkg", "schema"},
	}
	for _, test := range tests {
		t.Run(test.pkg, func(t *testing.T) {
			scope, err := manifest.Classify(test.pkg)
			if err != nil {
				t.Fatalf("Classify(%q) returned error: %v", test.pkg, err)
			}
			if scope != test.wantScope {
				t.Fatalf("Classify(%q) = %q, want %q", test.pkg, scope, test.wantScope)
			}
		})
	}

	// Omitted-scope negative fixture: a changed package with no manifest entry.
	if _, err := manifest.Classify("internal/unlisted"); err == nil {
		t.Fatalf("Classify omitted package = nil error, want missing error")
	} else if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("Classify omitted package error %q does not contain %q", err, "missing")
	}
}

func TestParseProfile(t *testing.T) {
	profile := `mode: set
github.com/unbound-force/unbound-force/internal/schemas/types.go:10.2,12.4 1 0
github.com/unbound-force/unbound-force/internal/schemas/registry.go:5.1,6.2 9 9
github.com/unbound-force/unbound-force/internal/doctor/checks.go:1.1,2.2 2 0
`
	coverage, err := ParseProfile(strings.NewReader(profile))
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	// schemas: 1+9=10 total, 0+9=9 covered -> 90%.
	if got := coverage["github.com/unbound-force/unbound-force/internal/schemas"]; got != 90 {
		t.Fatalf("schemas coverage = %v, want 90", got)
	}
	// doctor: 2 total, 0 covered -> 0%.
	if got := coverage["github.com/unbound-force/unbound-force/internal/doctor"]; got != 0 {
		t.Fatalf("doctor coverage = %v, want 0", got)
	}
}

func TestParseProfile_RejectsMalformed(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{"too few fields", "mode: set\nfile.go:1.1,2.2 3\n"},
		{"bad statement count", "mode: set\nfile.go:1.1,2.2 nope 1\n"},
		{"bad execution count", "mode: set\nfile.go:1.1,2.2 3 nope\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseProfile(strings.NewReader(test.doc)); err == nil {
				t.Fatalf("ParseProfile(%q) = nil error, want error", test.name)
			}
		})
	}
}

func TestPackagesFromFiles(t *testing.T) {
	files := []string{
		"internal/scaffold/scaffold.go",
		"internal/scaffold/scaffold_test.go",
		"internal/scaffold/assets/embed.go",
		"internal/doctor/checks.go",
		"internal/doctor/checks_test.go",
		"vendor/example/vendor.go",
		"internal/example/zz_generated.go",
		"README.md",
		"internal/doctor/checks.go",
	}
	got := PackagesFromFiles(files)
	want := []string{"internal/doctor", "internal/scaffold", "internal/scaffold/assets"}
	if len(got) != len(want) {
		t.Fatalf("PackagesFromFiles = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("PackagesFromFiles = %v, want %v", got, want)
		}
	}
}

func TestCheck(t *testing.T) {
	manifest, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	coverage := map[string]float64{
		"internal/scaffold":  85,
		"internal/artifacts": 95,
		"internal/schemas":   87,
		"internal/artifactpkg": 85,
	}

	t.Run("passes when all changed packages meet thresholds", func(t *testing.T) {
		failures, err := Check(CheckOptions{
			Manifest:        manifest,
			ChangedPackages: []string{"internal/scaffold", "internal/artifacts", "internal/schemas"},
			Coverage:        coverage,
		})
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if len(failures) != 0 {
			t.Fatalf("Check failures = %v, want none", failures)
		}
	})

	t.Run("fails a package below its scope threshold", func(t *testing.T) {
		failures, err := Check(CheckOptions{
			Manifest:        manifest,
			ChangedPackages: []string{"internal/artifactpkg"},
			Coverage:        coverage,
		})
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if len(failures) != 1 || !strings.Contains(failures[0], "below") {
			t.Fatalf("Check failures = %v, want one below-threshold failure", failures)
		}
	})

	t.Run("fails a changed package missing from the manifest", func(t *testing.T) {
		failures, err := Check(CheckOptions{
			Manifest:        manifest,
			ChangedPackages: []string{"internal/unlisted"},
			Coverage:        coverage,
		})
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if len(failures) != 1 || !strings.Contains(failures[0], "missing") {
			t.Fatalf("Check failures = %v, want one missing failure", failures)
		}
	})

	t.Run("fails a changed package with no measured coverage", func(t *testing.T) {
		failures, err := Check(CheckOptions{
			Manifest:        manifest,
			ChangedPackages: []string{"internal/artifacts"},
			Coverage:        map[string]float64{},
		})
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if len(failures) != 1 || !strings.Contains(failures[0], "no measured coverage") {
			t.Fatalf("Check failures = %v, want one no-coverage failure", failures)
		}
	})

	t.Run("collapses duplicate changed packages", func(t *testing.T) {
		failures, err := Check(CheckOptions{
			Manifest:        manifest,
			ChangedPackages: []string{"internal/scaffold", "internal/scaffold"},
			Coverage:        coverage,
		})
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		if len(failures) != 0 {
			t.Fatalf("Check failures = %v, want none for deduplicated input", failures)
		}
	})

	t.Run("rejects a nil manifest", func(t *testing.T) {
		if _, err := Check(CheckOptions{Manifest: nil}); err == nil {
			t.Fatalf("Check with nil manifest = nil error, want error")
		}
	})
}
