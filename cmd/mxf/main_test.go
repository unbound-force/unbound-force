package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/coaching"
	"github.com/unbound-force/unbound-force/internal/impediment"
	"github.com/unbound-force/unbound-force/internal/metrics"
	"github.com/unbound-force/unbound-force/internal/sprint"
)

// stubGHRunner satisfies sync.GHRunner for CLI-level tests.
type stubGHRunner struct {
	out []byte
	err error
}

func (s *stubGHRunner) Run(args ...string) ([]byte, error) {
	return s.out, s.err
}

func TestMxFParams_Defaults(t *testing.T) {
	p := &MxFParams{}
	p.defaults()

	if p.DataDir != ".mx-f" {
		t.Errorf("DataDir = %q, want .mx-f", p.DataDir)
	}
	if p.Stdout == nil {
		t.Error("Stdout should not be nil after defaults()")
	}
	if p.Stderr == nil {
		t.Error("Stderr should not be nil after defaults()")
	}
	if p.Now == nil {
		t.Error("Now should not be nil after defaults()")
	}
	if p.GHRunner == nil {
		t.Error("GHRunner should not be nil after defaults()")
	}
}

func TestNewRootCmd_HasSubcommands(t *testing.T) {
	cmd := newRootCmd()

	expected := []string{
		"collect", "metrics", "impediment",
		"dashboard", "sprint", "standup", "retro",
	}
	found := make(map[string]bool)
	for _, c := range cmd.Commands() {
		found[c.Name()] = true
	}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestNewRootCmdWithParams_InjectsStubs(t *testing.T) {
	stub := &stubGHRunner{out: []byte("ok"), err: nil}
	var buf bytes.Buffer
	p := &MxFParams{
		Stdout:   &buf,
		Stderr:   &buf,
		GHRunner: stub,
		DataDir:  t.TempDir(),
		Now:      func() time.Time { return time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC) },
	}
	cmd := newRootCmdWithParams(p)
	if cmd == nil {
		t.Fatal("newRootCmdWithParams returned nil")
	}
}

func TestRunCollect_NoData(t *testing.T) {
	var buf bytes.Buffer
	p := MxFParams{
		DataDir: t.TempDir(),
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC) },
	}
	p.defaults()
	// Collecting with no repo and no artifacts should succeed
	// (graceful degradation — 0/4 sources is not an error)
	err := runCollect(p, "all", "", "30d")
	if err != nil {
		t.Errorf("expected graceful degradation, got error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "0/4 sources") {
		t.Errorf("expected 0/4 sources in output, got:\n%s", output)
	}
}

// --- helper to write a fixture snapshot to disk ---

func writeFixtureSnapshot(t *testing.T, dataDir string, snap metrics.MetricsSnapshot) {
	t.Helper()
	store := metrics.NewStore(dataDir)
	if err := store.WriteSnapshot(snap); err != nil {
		t.Fatalf("writeFixtureSnapshot: %v", err)
	}
}

func fixtureSnapshot(ts time.Time) metrics.MetricsSnapshot {
	return metrics.MetricsSnapshot{
		Timestamp:        ts,
		Velocity:         12.5,
		CycleTime:        metrics.CycleTimeStats{Avg: 18.0, Median: 14.0, P90: 32.0, P99: 48.0},
		LeadTime:         72.0,
		DefectRate:       0.05,
		ReviewIterations: 1.8,
		CIPassRate:       95.0,
		BacklogHealth:    metrics.BacklogHealth{Total: 20, Ready: 16, Stale: 1},
		FlowEfficiency:   78.0,
		SourcesCollected: []string{"github", "gaze"},
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
}

// =============================================================
// runMetrics tests (decompose_and_test: skeleton + key paths)
// =============================================================

func TestRunMetrics_SummaryText(t *testing.T) {
	dataDir := t.TempDir()
	snap := fixtureSnapshot(fixedNow().Add(-1 * time.Hour))
	writeFixtureSnapshot(t, dataDir, snap)

	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dataDir,
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return fixedNow() },
	}
	p.defaults()

	if err := runMetrics(p, "summary", "text", 0, "30d"); err != nil {
		t.Fatalf("runMetrics summary: %v", err)
	}
	out := buf.String()

	for _, want := range []string{"Metrics Summary", "12.5", "18.0", "95.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary output missing %q, got:\n%s", want, out)
		}
	}
}

func TestRunMetrics_SummaryJSON(t *testing.T) {
	dataDir := t.TempDir()
	snap := fixtureSnapshot(fixedNow().Add(-1 * time.Hour))
	writeFixtureSnapshot(t, dataDir, snap)

	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dataDir,
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return fixedNow() },
	}
	p.defaults()

	if err := runMetrics(p, "summary", "json", 0, "30d"); err != nil {
		t.Fatalf("runMetrics summary json: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if v, ok := result["velocity"].(float64); !ok || v != 12.5 {
		t.Errorf("velocity = %v, want 12.5", result["velocity"])
	}
}

func TestRunMetrics_VelocityText(t *testing.T) {
	dataDir := t.TempDir()
	// Write two snapshots so we get two velocity points.
	writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
		Timestamp: fixedNow().Add(-48 * time.Hour),
		Velocity:  8.0,
	})
	writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
		Timestamp: fixedNow().Add(-1 * time.Hour),
		Velocity:  12.5,
	})

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runMetrics(p, "velocity", "text", 0, "30d"); err != nil {
		t.Fatalf("runMetrics velocity: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Sprint 1") || !strings.Contains(out, "Sprint 2") {
		t.Errorf("velocity output missing sprint labels, got:\n%s", out)
	}
	if !strings.Contains(out, "12.5") {
		t.Errorf("velocity output missing value 12.5, got:\n%s", out)
	}
}

func TestRunMetrics_CycleTimeText(t *testing.T) {
	dataDir := t.TempDir()
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runMetrics(p, "cycle-time", "text", 0, "30d"); err != nil {
		t.Fatalf("runMetrics cycle-time: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Cycle Time") {
		t.Errorf("cycle-time output missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "18.0") {
		t.Errorf("cycle-time output missing avg value, got:\n%s", out)
	}
}

func TestRunMetrics_BottlenecksText(t *testing.T) {
	dataDir := t.TempDir()
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runMetrics(p, "bottlenecks", "text", 0, "30d"); err != nil {
		t.Fatalf("runMetrics bottlenecks: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Bottleneck Analysis") {
		t.Errorf("bottlenecks output missing header, got:\n%s", out)
	}
	// Should contain at least one stage name
	if !strings.Contains(out, "Review") && !strings.Contains(out, "Testing") {
		t.Errorf("bottlenecks output missing stage names, got:\n%s", out)
	}
}

func TestRunMetrics_HealthText(t *testing.T) {
	dataDir := t.TempDir()
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runMetrics(p, "health", "text", 0, "30d"); err != nil {
		t.Fatalf("runMetrics health: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Health") {
		t.Errorf("health output missing header, got:\n%s", out)
	}
}

func TestRunMetrics_UnknownSubcommand(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	err := runMetrics(p, "nonexistent", "text", 0, "30d")
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown metrics subcommand") {
		t.Errorf("error = %q, want 'unknown metrics subcommand'", err.Error())
	}
}

// TODO(decompose): runMetrics "velocity" JSON format test
// TODO(decompose): runMetrics "bottlenecks" JSON format test
// TODO(decompose): runMetrics "health" JSON format test
// TODO(decompose): runMetrics with --sprints flag limiting velocity output

// =============================================================
// runRetro tests (decompose_and_test: skeleton + key paths)
// =============================================================

func TestRunRetro_StartCreatesFile(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dataDir,
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return fixedNow() },
	}
	p.defaults()

	if err := runRetro(p, "start", ""); err != nil {
		t.Fatalf("runRetro start: %v", err)
	}

	// Verify output mentions the date
	out := buf.String()
	if !strings.Contains(out, "2026-03-20") {
		t.Errorf("start output missing date, got:\n%s", out)
	}
	if !strings.Contains(out, "Retrospective started") {
		t.Errorf("start output missing confirmation, got:\n%s", out)
	}

	// Verify the retro file was created on disk
	retroFile := filepath.Join(dataDir, "retros", "2026-03-20-retro.md")
	if _, err := os.Stat(retroFile); os.IsNotExist(err) {
		t.Errorf("expected retro file at %s, but it does not exist", retroFile)
	}
}

func TestRunRetro_StartWithMetrics(t *testing.T) {
	dataDir := t.TempDir()
	// Write fixture metrics so retro start can include data
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dataDir,
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return fixedNow() },
	}
	p.defaults()

	if err := runRetro(p, "start", ""); err != nil {
		t.Fatalf("runRetro start with metrics: %v", err)
	}

	// Verify file was created
	retroFile := filepath.Join(dataDir, "retros", "2026-03-20-retro.md")
	data, err := os.ReadFile(retroFile)
	if err != nil {
		t.Fatalf("read retro file: %v", err)
	}
	content := string(data)
	// The file should contain the metrics data in its frontmatter
	if !strings.Contains(content, "velocity") {
		t.Errorf("retro file missing metrics data, got:\n%s", content)
	}
}

func TestRunRetro_ActionsListsItems(t *testing.T) {
	dataDir := t.TempDir()
	retroDir := filepath.Join(dataDir, "retros")

	// Create a retro with action items via the domain API
	retroStore := coaching.NewRetroStore(retroDir)
	record, err := retroStore.StartRetro("2026-03-15", nil)
	if err != nil {
		t.Fatalf("start retro fixture: %v", err)
	}
	record.ActionItems = []coaching.ActionItem{
		{
			ID:          "AI-001",
			Description: "Improve CI pipeline speed",
			Owner:       "cobalt",
			Deadline:    "2026-04-01",
			Status:      "pending",
		},
		{
			ID:          "AI-002",
			Description: "Add integration tests",
			Owner:       "gaze",
			Deadline:    "2026-04-15",
			Status:      "pending",
		},
	}
	if err := retroStore.SaveRetro(record); err != nil {
		t.Fatalf("save retro fixture: %v", err)
	}

	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dataDir,
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return fixedNow() },
	}
	p.defaults()

	if err := runRetro(p, "actions", "all"); err != nil {
		t.Fatalf("runRetro actions: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "AI-001") {
		t.Errorf("actions output missing AI-001, got:\n%s", out)
	}
	if !strings.Contains(out, "Improve CI pipeline speed") {
		t.Errorf("actions output missing description, got:\n%s", out)
	}
	if !strings.Contains(out, "AI-002") {
		t.Errorf("actions output missing AI-002, got:\n%s", out)
	}
}

func TestRunRetro_ActionsNoRetros(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dataDir,
		Stdout:  &buf,
		Stderr:  &buf,
		Now:     func() time.Time { return fixedNow() },
	}
	p.defaults()

	if err := runRetro(p, "actions", "all"); err != nil {
		t.Fatalf("runRetro actions (empty): %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "No action items found") {
		t.Errorf("expected 'No action items found', got:\n%s", out)
	}
}

func TestRunRetro_UnknownSubcommand(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	err := runRetro(p, "bogus", "")
	if err == nil {
		t.Fatal("expected error for unknown retro subcommand")
	}
	if !strings.Contains(err.Error(), "unknown retro subcommand") {
		t.Errorf("error = %q, want 'unknown retro subcommand'", err.Error())
	}
}

// TODO(decompose): runRetro "actions" with status filter "pending"
// TODO(decompose): runRetro "start" with previous action items review output

// =============================================================
// runDashboard tests (add_tests)
// =============================================================

func TestRunDashboard_ASCIIFullDashboard(t *testing.T) {
	dataDir := t.TempDir()
	// Write two snapshots for velocity chart content
	writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
		Timestamp: fixedNow().Add(-48 * time.Hour),
		Velocity:  8.0,
		CycleTime: metrics.CycleTimeStats{Avg: 20.0},
	})
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runDashboard(p, "", false, ""); err != nil {
		t.Fatalf("runDashboard ASCII: %v", err)
	}

	out := buf.String()
	// The full dashboard renders velocity bar chart + health indicators.
	// Bar chart uses "█" characters.
	if !strings.Contains(out, "█") {
		t.Errorf("ASCII dashboard missing bar chart characters, got:\n%s", out)
	}
	if !strings.Contains(out, "Health") {
		t.Errorf("ASCII dashboard missing Health section, got:\n%s", out)
	}
}

func TestRunDashboard_VelocitySubcommand(t *testing.T) {
	dataDir := t.TempDir()
	writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
		Timestamp: fixedNow().Add(-48 * time.Hour),
		Velocity:  8.0,
	})
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runDashboard(p, "velocity", false, ""); err != nil {
		t.Fatalf("runDashboard velocity: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Velocity") {
		t.Errorf("velocity dashboard missing title, got:\n%s", out)
	}
	if !strings.Contains(out, "█") {
		t.Errorf("velocity dashboard missing bar chars, got:\n%s", out)
	}
}

func TestRunDashboard_CycleTimeSparkline(t *testing.T) {
	dataDir := t.TempDir()
	// Write several snapshots so sparkline has multiple data points
	for i := 0; i < 5; i++ {
		writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
			Timestamp: fixedNow().Add(time.Duration(-5+i) * 24 * time.Hour),
			Velocity:  10.0,
			CycleTime: metrics.CycleTimeStats{Avg: float64(10 + i*3)},
		})
	}

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runDashboard(p, "cycle-time", false, ""); err != nil {
		t.Fatalf("runDashboard cycle-time: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Cycle Time") {
		t.Errorf("cycle-time dashboard missing title, got:\n%s", out)
	}
	// Sparkline uses Unicode block chars like ▁▂▃▄▅▆▇█
	hasSparkChar := false
	for _, r := range out {
		if r >= '▁' && r <= '█' {
			hasSparkChar = true
			break
		}
	}
	if !hasSparkChar {
		t.Errorf("cycle-time dashboard missing sparkline characters, got:\n%s", out)
	}
}

func TestRunDashboard_HealthSubcommand(t *testing.T) {
	dataDir := t.TempDir()
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runDashboard(p, "health", false, ""); err != nil {
		t.Fatalf("runDashboard health: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Health") {
		t.Errorf("health dashboard missing title, got:\n%s", out)
	}
	// Health indicators include dimension names
	if !strings.Contains(out, "velocity") {
		t.Errorf("health dashboard missing velocity indicator, got:\n%s", out)
	}
}

func TestRunDashboard_HTMLCreatesFile(t *testing.T) {
	dataDir := t.TempDir()
	writeFixtureSnapshot(t, dataDir, fixtureSnapshot(fixedNow().Add(-1*time.Hour)))

	htmlPath := filepath.Join(t.TempDir(), "test-dashboard.html")

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runDashboard(p, "", true, htmlPath); err != nil {
		t.Fatalf("runDashboard HTML: %v", err)
	}

	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read HTML dashboard: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<!DOCTYPE html>") {
		t.Errorf("HTML dashboard missing doctype, got:\n%.200s", content)
	}
	if !strings.Contains(content, "Mx F") {
		t.Errorf("HTML dashboard missing title, got:\n%.200s", content)
	}
	if !strings.Contains(content, "12.5") {
		t.Errorf("HTML dashboard missing velocity value, got:\n%.500s", content)
	}
}

func TestRunDashboard_NoData(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	err := runDashboard(p, "", false, "")
	if err == nil {
		t.Fatal("expected error with no data")
	}
	if !strings.Contains(err.Error(), "no metrics data") {
		t.Errorf("error = %q, want 'no metrics data'", err.Error())
	}
}

// =============================================================
// runSprint tests (add_tests)
// =============================================================

func TestRunSprint_PlanCreatesFile(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runSprint(p, "plan", "Ship v2.0 release"); err != nil {
		t.Fatalf("runSprint plan: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Sprint Planning") {
		t.Errorf("plan output missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "Ship v2.0 release") {
		t.Errorf("plan output missing goal, got:\n%s", out)
	}

	// Verify sprint JSON file was created
	sprintsDir := filepath.Join(dataDir, "sprints")
	entries, err := os.ReadDir(sprintsDir)
	if err != nil {
		t.Fatalf("read sprints dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no sprint file created")
	}
	// Read and verify JSON content
	data, err := os.ReadFile(filepath.Join(sprintsDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("read sprint file: %v", err)
	}
	var state sprint.SprintState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("parse sprint JSON: %v", err)
	}
	if state.Goal != "Ship v2.0 release" {
		t.Errorf("sprint goal = %q, want %q", state.Goal, "Ship v2.0 release")
	}
	if state.Status != "active" {
		t.Errorf("sprint status = %q, want %q", state.Status, "active")
	}
}

func TestRunSprint_PlanWithHistoricalVelocity(t *testing.T) {
	dataDir := t.TempDir()
	// Write metrics snapshots to influence velocity calculation
	writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
		Timestamp: fixedNow().Add(-48 * time.Hour),
		Velocity:  10.0,
	})
	writeFixtureSnapshot(t, dataDir, metrics.MetricsSnapshot{
		Timestamp: fixedNow().Add(-1 * time.Hour),
		Velocity:  14.0,
	})

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runSprint(p, "plan", ""); err != nil {
		t.Fatalf("runSprint plan: %v", err)
	}

	out := buf.String()
	// Average velocity should be (10+14)/2 = 12.0
	if !strings.Contains(out, "12.0") {
		t.Errorf("plan output missing historical velocity, got:\n%s", out)
	}
}

func TestRunSprint_ReviewNoSprint(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	err := runSprint(p, "review", "")
	if err == nil {
		t.Fatal("expected error when reviewing with no sprint")
	}
	if !strings.Contains(err.Error(), "no sprint data found") {
		t.Errorf("error = %q, want 'no sprint data found'", err.Error())
	}
}

func TestRunSprint_UnknownSubcommand(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	err := runSprint(p, "bogus", "")
	if err == nil {
		t.Fatal("expected error for unknown sprint subcommand")
	}
	if !strings.Contains(err.Error(), "unknown sprint subcommand") {
		t.Errorf("error = %q, want 'unknown sprint subcommand'", err.Error())
	}
}

// =============================================================
// runStandup tests (add_tests)
// =============================================================

func TestRunStandup_NoActiveSprint(t *testing.T) {
	dataDir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runStandup(p); err != nil {
		t.Fatalf("runStandup: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Daily Standup") {
		t.Errorf("standup output missing header, got:\n%s", out)
	}
	if !strings.Contains(out, "No active sprint") {
		t.Errorf("standup output missing 'No active sprint', got:\n%s", out)
	}
}

func TestRunStandup_WithSprintAndImpediments(t *testing.T) {
	dataDir := t.TempDir()

	// Create a sprint via domain API
	sprintStore := sprint.NewSprintStore(filepath.Join(dataDir, "sprints"))
	_, err := sprintStore.Plan("Deliver alpha milestone", 10.0, []string{"item-1", "item-2"})
	if err != nil {
		t.Fatalf("create sprint fixture: %v", err)
	}

	// Create impediments via domain API
	impRepo := impediment.NewRepository(filepath.Join(dataDir, "impediments"))
	_, err = impRepo.Add("CI flaky tests", "high", "gaze", "Tests fail intermittently", fixedNow())
	if err != nil {
		t.Fatalf("create impediment fixture: %v", err)
	}
	_, err = impRepo.Add("Blocked on upstream API", "critical", "cobalt", "Waiting for v2 API", fixedNow())
	if err != nil {
		t.Fatalf("create impediment fixture 2: %v", err)
	}

	var buf bytes.Buffer
	p := MxFParams{DataDir: dataDir, Stdout: &buf, Stderr: &buf, Now: func() time.Time { return fixedNow() }}
	p.defaults()

	if err := runStandup(p); err != nil {
		t.Fatalf("runStandup: %v", err)
	}

	out := buf.String()
	// Verify sprint info is present
	if !strings.Contains(out, "sprint-") {
		t.Errorf("standup output missing sprint name, got:\n%s", out)
	}
	if !strings.Contains(out, "Planned: 2 items") {
		t.Errorf("standup output missing planned items count, got:\n%s", out)
	}

	// Verify blocked items are listed
	if !strings.Contains(out, "Blocked") {
		t.Errorf("standup output missing Blocked section, got:\n%s", out)
	}
	if !strings.Contains(out, "CI flaky tests") {
		t.Errorf("standup output missing impediment title, got:\n%s", out)
	}
	if !strings.Contains(out, "critical") {
		t.Errorf("standup output missing impediment severity, got:\n%s", out)
	}
}

// =============================================================
// runImpediment tests (add_tests)
// =============================================================

func TestRunImpedimentAdd_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{
		DataDir: dir, Stdout: &buf, Stderr: &buf,
		Now: func() time.Time { return time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC) },
	}
	p.defaults()
	err := runImpedimentAdd(p, "Flaky CI", "high", "@dev", "CI is flaky")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IMP-001") {
		t.Errorf("expected IMP-001 in output, got: %s", buf.String())
	}
}

func TestRunImpedimentList_TextFormat(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	now := func() time.Time { return time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC) }
	p := MxFParams{DataDir: dir, Stdout: &buf, Stderr: &buf, Now: now}
	p.defaults()
	// Add impediments first
	_ = runImpedimentAdd(p, "High issue", "high", "@dev", "desc")
	buf.Reset()
	_ = runImpedimentAdd(p, "Low issue", "low", "@dev2", "desc2")
	buf.Reset()

	err := runImpedimentList(p, "all", "text")
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "IMP-001") {
		t.Error("missing IMP-001")
	}
	if !strings.Contains(output, "IMP-002") {
		t.Error("missing IMP-002")
	}
}

func TestRunImpedimentList_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	now := func() time.Time { return time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC) }
	p := MxFParams{DataDir: dir, Stdout: &buf, Stderr: &buf, Now: now}
	p.defaults()
	_ = runImpedimentAdd(p, "Test", "medium", "@dev", "desc")
	buf.Reset()

	err := runImpedimentList(p, "all", "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IMP-001") {
		t.Error("JSON should contain IMP-001")
	}
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "[") {
		t.Error("JSON output should be an array")
	}
}

func TestRunImpedimentList_Empty(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dir, Stdout: &buf, Stderr: &buf}
	p.defaults()
	err := runImpedimentList(p, "all", "text")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No impediments") {
		t.Errorf("expected 'No impediments' message, got: %s", buf.String())
	}
}

func TestRunImpedimentDetect_NoData(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	p := MxFParams{DataDir: dir, Stdout: &buf, Stderr: &buf}
	p.defaults()
	// With no metrics data, detect should report insufficient data
	err := runImpedimentDetect(p)
	if err == nil {
		// Either error or "No potential impediments" message is acceptable
		if !strings.Contains(buf.String(), "No potential") && !strings.Contains(buf.String(), "insufficient") {
			t.Errorf("expected insufficient data message, got: %s", buf.String())
		}
	}
}
