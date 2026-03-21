package orchestration

import (
	"path/filepath"
	"testing"
	"time"
)

func TestWorkflowStore_SaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := &WorkflowStore{Dir: dir}

	now := time.Date(2026, 3, 20, 14, 30, 0, 0, time.UTC)
	wf := &WorkflowInstance{
		WorkflowID:    "wf-test-001",
		FeatureBranch: "feat/health-check",
		BacklogItemID: "BI-042",
		Stages: []WorkflowStage{
			{StageName: StageDefine, Hero: "muti-mind", Status: StatusPending},
			{StageName: StageImplement, Hero: "cobalt-crush", Status: StatusPending},
		},
		CurrentStage:    0,
		StartedAt:       now,
		Status:          StatusActive,
		AvailableHeroes: []string{"muti-mind", "cobalt-crush"},
		IterationCount:  0,
	}

	if err := store.Save(wf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load("wf-test-001")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.WorkflowID != "wf-test-001" {
		t.Errorf("WorkflowID = %q, want %q", loaded.WorkflowID, "wf-test-001")
	}
	if loaded.FeatureBranch != "feat/health-check" {
		t.Errorf("FeatureBranch = %q, want %q", loaded.FeatureBranch, "feat/health-check")
	}
	if loaded.BacklogItemID != "BI-042" {
		t.Errorf("BacklogItemID = %q, want %q", loaded.BacklogItemID, "BI-042")
	}
	if loaded.Status != StatusActive {
		t.Errorf("Status = %q, want %q", loaded.Status, StatusActive)
	}
	if len(loaded.Stages) != 2 {
		t.Errorf("len(Stages) = %d, want 2", len(loaded.Stages))
	}
	if !loaded.StartedAt.Equal(now) {
		t.Errorf("StartedAt = %v, want %v", loaded.StartedAt, now)
	}
}

func TestWorkflowStore_List_FilterByStatus(t *testing.T) {
	dir := t.TempDir()
	store := &WorkflowStore{Dir: dir}

	now := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)

	// Save 3 workflows with different statuses
	for i, status := range []string{StatusActive, StatusCompleted, StatusActive} {
		wf := &WorkflowInstance{
			WorkflowID:    "wf-list-" + string(rune('a'+i)),
			FeatureBranch: "feat/test",
			StartedAt:     now.Add(time.Duration(i) * time.Hour),
			Status:        status,
		}
		if err := store.Save(wf); err != nil {
			t.Fatalf("Save wf-%c failed: %v", rune('a'+i), err)
		}
	}

	// Filter by active
	active, err := store.List(StatusActive)
	if err != nil {
		t.Fatalf("List(active) failed: %v", err)
	}
	if len(active) != 2 {
		t.Errorf("expected 2 active workflows, got %d", len(active))
	}

	// Filter by completed
	completed, err := store.List(StatusCompleted)
	if err != nil {
		t.Fatalf("List(completed) failed: %v", err)
	}
	if len(completed) != 1 {
		t.Errorf("expected 1 completed workflow, got %d", len(completed))
	}

	// List all
	all, err := store.List("")
	if err != nil {
		t.Fatalf("List('') failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 total workflows, got %d", len(all))
	}

	// Verify descending sort by started_at
	if all[0].StartedAt.Before(all[1].StartedAt) {
		t.Error("expected workflows sorted by started_at descending")
	}
}

func TestWorkflowStore_Latest_ByBranch(t *testing.T) {
	dir := t.TempDir()
	store := &WorkflowStore{Dir: dir}

	now := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)

	// Two active workflows on different branches
	wf1 := &WorkflowInstance{
		WorkflowID:    "wf-branch-a",
		FeatureBranch: "feat/alpha",
		StartedAt:     now,
		Status:        StatusActive,
	}
	wf2 := &WorkflowInstance{
		WorkflowID:    "wf-branch-b",
		FeatureBranch: "feat/beta",
		StartedAt:     now.Add(time.Hour),
		Status:        StatusActive,
	}

	if err := store.Save(wf1); err != nil {
		t.Fatalf("Save wf1 failed: %v", err)
	}
	if err := store.Save(wf2); err != nil {
		t.Fatalf("Save wf2 failed: %v", err)
	}

	latest, err := store.Latest("feat/alpha")
	if err != nil {
		t.Fatalf("Latest failed: %v", err)
	}
	if latest == nil {
		t.Fatal("expected non-nil latest for feat/alpha")
	}
	if latest.WorkflowID != "wf-branch-a" {
		t.Errorf("WorkflowID = %q, want %q", latest.WorkflowID, "wf-branch-a")
	}

	// Non-existent branch
	missing, err := store.Latest("feat/nonexistent")
	if err != nil {
		t.Fatalf("Latest for nonexistent failed: %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for nonexistent branch, got %+v", missing)
	}
}

func TestWorkflowStore_Load_MissingFile(t *testing.T) {
	dir := t.TempDir()
	store := &WorkflowStore{Dir: dir}

	_, err := store.Load("nonexistent-workflow")
	if err == nil {
		t.Error("expected error for missing workflow file")
	}
}

func TestWorkflowStore_List_Empty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty-workflows")
	store := &WorkflowStore{Dir: dir}

	workflows, err := store.List("")
	if err != nil {
		t.Fatalf("List on empty dir failed: %v", err)
	}
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows, got %d", len(workflows))
	}
}
