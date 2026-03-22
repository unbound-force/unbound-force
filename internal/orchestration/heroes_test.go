package orchestration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// stubLookPath returns a function that simulates exec.LookPath.
// Binaries in the found set return a path; others return an error.
func stubLookPath(found map[string]bool) func(string) (string, error) {
	return func(name string) (string, error) {
		if found[name] {
			return "/usr/local/bin/" + name, nil
		}
		return "", fmt.Errorf("executable %q not found", name)
	}
}

func TestDetectHeroes_AllPresent(t *testing.T) {
	dir := t.TempDir()

	// Create all agent files
	for _, name := range []string{
		"muti-mind-po.md",
		"cobalt-crush-dev.md",
		"divisor-guard.md",
		"mx-f-coach.md",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# agent"), 0644); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	// All binaries available
	lookPath := stubLookPath(map[string]bool{"gaze": true, "mxf": true})

	heroes, err := DetectHeroes(dir, lookPath)
	if err != nil {
		t.Fatalf("DetectHeroes failed: %v", err)
	}

	if len(heroes) != 5 {
		t.Fatalf("expected 5 heroes, got %d", len(heroes))
	}

	for _, h := range heroes {
		if !h.Available {
			t.Errorf("hero %q should be available", h.Name)
		}
	}
}

func TestDetectHeroes_MissingHeroes(t *testing.T) {
	dir := t.TempDir()

	// Only create muti-mind and cobalt-crush agents
	for _, name := range []string{"muti-mind-po.md", "cobalt-crush-dev.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# agent"), 0644); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	// No binaries available
	lookPath := stubLookPath(map[string]bool{})

	heroes, err := DetectHeroes(dir, lookPath)
	if err != nil {
		t.Fatalf("DetectHeroes failed: %v", err)
	}

	available := make(map[string]bool)
	for _, h := range heroes {
		available[h.Name] = h.Available
	}

	if !available["muti-mind"] {
		t.Error("muti-mind should be available")
	}
	if !available["cobalt-crush"] {
		t.Error("cobalt-crush should be available")
	}
	if available["gaze"] {
		t.Error("gaze should NOT be available")
	}
	if available["divisor"] {
		t.Error("divisor should NOT be available")
	}
	if available["mx-f"] {
		t.Error("mx-f should NOT be available")
	}
}

func TestDetectHeroes_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	lookPath := stubLookPath(map[string]bool{})

	heroes, err := DetectHeroes(dir, lookPath)
	if err != nil {
		t.Fatalf("DetectHeroes failed: %v", err)
	}

	for _, h := range heroes {
		if h.Available {
			t.Errorf("hero %q should NOT be available in empty dir", h.Name)
		}
	}
}

func TestStageHeroMap(t *testing.T) {
	m := StageHeroMap()

	expected := map[string]string{
		StageDefine:    "muti-mind",
		StageImplement: "cobalt-crush",
		StageValidate:  "gaze",
		StageReview:    "divisor",
		StageAccept:    "muti-mind",
		StageReflect:   "mx-f",
	}

	for stage, hero := range expected {
		if m[stage] != hero {
			t.Errorf("StageHeroMap[%q] = %q, want %q", stage, m[stage], hero)
		}
	}
}

func TestStageExecutionModeMap(t *testing.T) {
	m := StageExecutionModeMap()

	if len(m) != 6 {
		t.Fatalf("StageExecutionModeMap has %d entries, want 6", len(m))
	}

	// Per FR-002: define=human, implement=swarm, validate=swarm,
	// review=swarm, accept=human, reflect=swarm.
	expected := map[string]string{
		StageDefine:    ModeHuman,
		StageImplement: ModeSwarm,
		StageValidate:  ModeSwarm,
		StageReview:    ModeSwarm,
		StageAccept:    ModeHuman,
		StageReflect:   ModeSwarm,
	}

	for stage, mode := range expected {
		if m[stage] != mode {
			t.Errorf("StageExecutionModeMap[%q] = %q, want %q", stage, m[stage], mode)
		}
	}
}
