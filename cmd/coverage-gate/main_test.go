package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/coveragegate"
)

func TestResolveBase(t *testing.T) {
	t.Run("explicit flag wins", func(t *testing.T) {
		t.Setenv("GITHUB_BASE_SHA", "env-sha")
		if got := resolveBase("flag-sha"); got != "flag-sha" {
			t.Fatalf("resolveBase = %q, want flag-sha", got)
		}
	})
	t.Run("falls back to environment", func(t *testing.T) {
		t.Setenv("GITHUB_BASE_SHA", "env-sha")
		if got := resolveBase(""); got != "env-sha" {
			t.Fatalf("resolveBase = %q, want env-sha", got)
		}
	})
	t.Run("falls back to main", func(t *testing.T) {
		t.Setenv("GITHUB_BASE_SHA", "")
		if got := resolveBase(""); got != "main" {
			t.Fatalf("resolveBase = %q, want main", got)
		}
	})
}

func TestModulePath(t *testing.T) {
	module, err := modulePath()
	if err != nil {
		t.Fatalf("modulePath: %v", err)
	}
	if module != "github.com/unbound-force/unbound-force" {
		t.Fatalf("modulePath = %q, want github.com/unbound-force/unbound-force", module)
	}
}

func TestLoadManifest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "coverage-gate.json")
		if err := os.WriteFile(path, []byte(`{"version":1,"scopes":{"helper":{"threshold":80,"packages":[]}}}`), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		manifest, err := loadManifest(path)
		if err != nil {
			t.Fatalf("loadManifest: %v", err)
		}
		if manifest.Version != 1 {
			t.Fatalf("version = %d, want 1", manifest.Version)
		}
	})
	t.Run("missing", func(t *testing.T) {
		if _, err := loadManifest(filepath.Join(t.TempDir(), "nope.json")); err == nil {
			t.Fatalf("loadManifest of missing file = nil error, want error")
		}
	})
}

func TestRunGate(t *testing.T) {
	manifest, err := coveragegate.ParseManifest([]byte(`{"version":1,"scopes":{"helper":{"threshold":80,"packages":["internal/scaffold"]}}}`))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}

	t.Run("succeeds within thresholds", func(t *testing.T) {
		var stdout, stderr strings.Builder
		if code := runGate(manifest, []string{"internal/scaffold"}, map[string]float64{"internal/scaffold": 90}, &stdout, &stderr); code != 0 {
			t.Fatalf("runGate = %d, stderr %s, want 0", code, stderr.String())
		}
	})

	t.Run("fails below threshold", func(t *testing.T) {
		var stdout, stderr strings.Builder
		if code := runGate(manifest, []string{"internal/scaffold"}, map[string]float64{"internal/scaffold": 60}, &stdout, &stderr); code == 0 {
			t.Fatalf("runGate = 0, want non-zero")
		}
		if !strings.Contains(stderr.String(), "below") {
			t.Fatalf("stderr %q missing below-threshold detail", stderr.String())
		}
	})

	t.Run("fails missing package", func(t *testing.T) {
		var stdout, stderr strings.Builder
		if code := runGate(manifest, []string{"internal/unlisted"}, map[string]float64{}, &stdout, &stderr); code == 0 {
			t.Fatalf("runGate = 0, want non-zero")
		}
		if !strings.Contains(stderr.String(), "missing") {
			t.Fatalf("stderr %q missing missing-package detail", stderr.String())
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("succeeds with no changed packages", func(t *testing.T) {
		dir := t.TempDir()
		manifestPath := filepath.Join(dir, "coverage-gate.json")
		profilePath := filepath.Join(dir, "coverage.out")
		if err := os.WriteFile(manifestPath, []byte(`{"version":1,"scopes":{"helper":{"threshold":80,"packages":[]}}}`), 0o600); err != nil {
			t.Fatalf("WriteFile manifest: %v", err)
		}
		if err := os.WriteFile(profilePath, []byte("mode: set\n"), 0o600); err != nil {
			t.Fatalf("WriteFile profile: %v", err)
		}
		var stdout, stderr strings.Builder
		if code := run([]string{"-manifest", manifestPath, "-profile", profilePath, "-base", "HEAD"}, &stdout, &stderr); code != 0 {
			t.Fatalf("run exit = %d, stderr = %s, want 0", code, stderr.String())
		}
	})

	t.Run("fails on missing manifest", func(t *testing.T) {
		var stdout, stderr strings.Builder
		if code := run([]string{"-manifest", filepath.Join(t.TempDir(), "nope.json")}, &stdout, &stderr); code == 0 {
			t.Fatalf("run exit = 0, want non-zero; stderr = %s", stderr.String())
		}
	})

	t.Run("fails on non-numeric threshold", func(t *testing.T) {
		dir := t.TempDir()
		manifestPath := filepath.Join(dir, "coverage-gate.json")
		profilePath := filepath.Join(dir, "coverage.out")
		if err := os.WriteFile(manifestPath, []byte(`{"version":1,"scopes":{"helper":{"threshold":"eighty","packages":[]}}}`), 0o600); err != nil {
			t.Fatalf("WriteFile manifest: %v", err)
		}
		if err := os.WriteFile(profilePath, []byte("mode: set\n"), 0o600); err != nil {
			t.Fatalf("WriteFile profile: %v", err)
		}
		var stdout, stderr strings.Builder
		if code := run([]string{"-manifest", manifestPath, "-profile", profilePath, "-base", "HEAD"}, &stdout, &stderr); code == 0 {
			t.Fatalf("run exit = 0, want non-zero; stderr = %s", stderr.String())
		}
	})

	t.Run("fails on missing coverage profile", func(t *testing.T) {
		dir := t.TempDir()
		manifestPath := filepath.Join(dir, "coverage-gate.json")
		if err := os.WriteFile(manifestPath, []byte(`{"version":1,"scopes":{"helper":{"threshold":80,"packages":[]}}}`), 0o600); err != nil {
			t.Fatalf("WriteFile manifest: %v", err)
		}
		var stdout, stderr strings.Builder
		if code := run([]string{"-manifest", manifestPath, "-profile", filepath.Join(dir, "nope.out"), "-base", "HEAD"}, &stdout, &stderr); code == 0 {
			t.Fatalf("run exit = 0, want non-zero; stderr = %s", stderr.String())
		}
	})

	t.Run("fails on unresolvable base", func(t *testing.T) {
		dir := t.TempDir()
		manifestPath := filepath.Join(dir, "coverage-gate.json")
		profilePath := filepath.Join(dir, "coverage.out")
		if err := os.WriteFile(manifestPath, []byte(`{"version":1,"scopes":{"helper":{"threshold":80,"packages":[]}}}`), 0o600); err != nil {
			t.Fatalf("WriteFile manifest: %v", err)
		}
		if err := os.WriteFile(profilePath, []byte("mode: set\n"), 0o600); err != nil {
			t.Fatalf("WriteFile profile: %v", err)
		}
		var stdout, stderr strings.Builder
		if code := run([]string{"-manifest", manifestPath, "-profile", profilePath, "-base", "refs/does/not/exist"}, &stdout, &stderr); code == 0 {
			t.Fatalf("run exit = 0, want non-zero; stderr = %s", stderr.String())
		}
	})
}
