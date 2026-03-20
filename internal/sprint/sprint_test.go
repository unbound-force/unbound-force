package sprint

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/impediment"
	"github.com/unbound-force/unbound-force/internal/metrics"
)

func TestSprintStore_PlanAndReview(t *testing.T) {
	dir := t.TempDir()
	store := NewSprintStore(dir)

	items := []string{"BI-001", "BI-002", "BI-003"}
	state, err := store.Plan("Test sprint", 10.0, items)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if state.Status != "active" {
		t.Errorf("Status = %q, want active", state.Status)
	}
	if len(state.PlannedItems) != 3 {
		t.Errorf("PlannedItems = %d, want 3", len(state.PlannedItems))
	}

	// Simulate completing items
	state.CompletedItems = []string{"BI-001", "BI-002"}
	_ = store.Save(state)

	reviewed, err := store.Review(state.SprintName)
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if reviewed.Status != "complete" {
		t.Errorf("Status = %q, want complete", reviewed.Status)
	}
	if reviewed.Velocity != 2.0 {
		t.Errorf("Velocity = %f, want 2.0", reviewed.Velocity)
	}
}

func TestSprintPlan_CapacityCalculation(t *testing.T) {
	dir := t.TempDir()
	store := NewSprintStore(dir)

	// More items than velocity allows
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("BI-%03d", i+1)
	}

	state, err := store.Plan("Capacity test", 10.0, items)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(state.PlannedItems) != 10 {
		t.Errorf("PlannedItems = %d, want 10 (capped by velocity)", len(state.PlannedItems))
	}
}

func TestSprintStore_Latest_Empty(t *testing.T) {
	dir := t.TempDir()
	store := NewSprintStore(dir)
	latest, err := store.Latest()
	if err != nil {
		t.Fatalf("Latest on empty dir: %v", err)
	}
	if latest != nil {
		t.Errorf("expected nil for empty dir, got %+v", latest)
	}
}

func TestSprintStore_Latest_MultipleSprints(t *testing.T) {
	dir := t.TempDir()
	store := NewSprintStore(dir)
	// Create two sprints — Latest should return the lexicographically latest
	s1 := &SprintState{SprintName: "sprint-2026-03-01", Goal: "first", Status: "complete"}
	s2 := &SprintState{SprintName: "sprint-2026-03-15", Goal: "second", Status: "active"}
	if err := store.Save(s1); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(s2); err != nil {
		t.Fatal(err)
	}

	latest, err := store.Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest == nil {
		t.Fatal("expected non-nil")
	}
	if latest.SprintName != "sprint-2026-03-15" {
		t.Errorf("SprintName = %q, want sprint-2026-03-15", latest.SprintName)
	}
}

func TestSprintStore_Load_MissingFile(t *testing.T) {
	dir := t.TempDir()
	store := NewSprintStore(dir)
	_, err := store.Load("nonexistent")
	if err == nil {
		t.Error("expected error for missing sprint file")
	}
}

func TestStandup_WithImpediments(t *testing.T) {
	sprintDir := t.TempDir()
	impDir := t.TempDir()
	metricsDir := t.TempDir()

	sprintStore := NewSprintStore(sprintDir)
	impRepo := impediment.NewRepository(impDir)
	metricsStore := metrics.NewStore(metricsDir)

	// Add an impediment
	now := time.Now()
	_, err := impRepo.Add("Blocked CI", "high", "@dev", "CI is broken", now)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = Standup(sprintStore, impRepo, metricsStore, &buf)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Blocked") {
		t.Errorf("expected blocked items in standup, got:\n%s", output)
	}
	if !strings.Contains(output, "IMP-001") {
		t.Error("expected impediment ID in standup")
	}
}
