package orchestration

import (
	"testing"
	"time"
)

func TestAnalyzeWorkflows_FrequentDivisorFindings(t *testing.T) {
	// 3 records with the same review error category
	records := []WorkflowRecord{
		{
			WorkflowID: "wf-001",
			Stages: []WorkflowStage{
				{StageName: StageReview, Status: StatusFailed, Error: "missing error wrapping"},
			},
		},
		{
			WorkflowID: "wf-002",
			Stages: []WorkflowStage{
				{StageName: StageReview, Status: StatusFailed, Error: "missing error wrapping"},
			},
		},
		{
			WorkflowID: "wf-003",
			Stages: []WorkflowStage{
				{StageName: StageReview, Status: StatusFailed, Error: "missing error wrapping"},
			},
		},
	}

	feedback, err := AnalyzeWorkflows(records, time.Now().UTC())
	if err != nil {
		t.Fatalf("AnalyzeWorkflows failed: %v", err)
	}

	if len(feedback) == 0 {
		t.Fatal("expected at least one feedback item for frequent findings")
	}

	// Verify feedback targets cobalt-crush
	found := false
	for _, fb := range feedback {
		if fb.TargetHero == "cobalt-crush" && fb.SourceHero == "divisor" {
			found = true
			if fb.PatternObserved == "" {
				t.Error("PatternObserved should not be empty")
			}
			if fb.Recommendation == "" {
				t.Error("Recommendation should not be empty")
			}
			if len(fb.WorkflowIDs) == 0 {
				t.Error("WorkflowIDs should not be empty")
			}
			break
		}
	}
	if !found {
		t.Error("expected feedback from divisor targeting cobalt-crush")
	}
}

func TestAnalyzeWorkflows_NoPatterns(t *testing.T) {
	// Only 2 records — below threshold
	records := []WorkflowRecord{
		{
			WorkflowID: "wf-001",
			Stages: []WorkflowStage{
				{StageName: StageReview, Status: StatusCompleted},
			},
		},
		{
			WorkflowID: "wf-002",
			Stages: []WorkflowStage{
				{StageName: StageReview, Status: StatusCompleted},
			},
		},
	}

	feedback, err := AnalyzeWorkflows(records, time.Now().UTC())
	if err != nil {
		t.Fatalf("AnalyzeWorkflows failed: %v", err)
	}

	if len(feedback) != 0 {
		t.Errorf("expected 0 feedback for 2 records (below threshold), got %d", len(feedback))
	}
}

func TestNextFeedbackID(t *testing.T) {
	tests := []struct {
		name     string
		existing []LearningFeedback
		expected string
	}{
		{
			name:     "no existing feedback",
			existing: nil,
			expected: "LF-001",
		},
		{
			name: "existing LF-003",
			existing: []LearningFeedback{
				{ID: "LF-001"},
				{ID: "LF-003"},
			},
			expected: "LF-004",
		},
		{
			name: "existing LF-010",
			existing: []LearningFeedback{
				{ID: "LF-010"},
			},
			expected: "LF-011",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextFeedbackID(tt.existing)
			if got != tt.expected {
				t.Errorf("NextFeedbackID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSaveFeedback_LoadFeedback_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	now := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)
	feedback := []LearningFeedback{
		{
			ID:              "LF-001",
			SourceHero:      "divisor",
			TargetHero:      "cobalt-crush",
			PatternObserved: "missing error wrapping in 3 workflows",
			Recommendation:  "update convention pack",
			SupportingData:  map[string]string{"count": "3"},
			Status:          "proposed",
			CreatedAt:       now,
			WorkflowIDs:     []string{"wf-001", "wf-002", "wf-003"},
		},
		{
			ID:              "LF-002",
			SourceHero:      "mx-f",
			TargetHero:      "cobalt-crush",
			PatternObserved: "velocity improving",
			Recommendation:  "continue current approach",
			Status:          "proposed",
			CreatedAt:       now,
			WorkflowIDs:     []string{"wf-001"},
		},
	}

	if err := SaveFeedback(dir, feedback); err != nil {
		t.Fatalf("SaveFeedback failed: %v", err)
	}

	loaded, err := LoadFeedback(dir)
	if err != nil {
		t.Fatalf("LoadFeedback failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 feedback items, got %d", len(loaded))
	}

	// Verify first item (sorted by ID)
	if loaded[0].ID != "LF-001" {
		t.Errorf("first item ID = %q, want %q", loaded[0].ID, "LF-001")
	}
	if loaded[0].SourceHero != "divisor" {
		t.Errorf("SourceHero = %q, want %q", loaded[0].SourceHero, "divisor")
	}
	if loaded[0].TargetHero != "cobalt-crush" {
		t.Errorf("TargetHero = %q, want %q", loaded[0].TargetHero, "cobalt-crush")
	}
	if len(loaded[0].WorkflowIDs) != 3 {
		t.Errorf("WorkflowIDs length = %d, want 3", len(loaded[0].WorkflowIDs))
	}

	// Verify second item
	if loaded[1].ID != "LF-002" {
		t.Errorf("second item ID = %q, want %q", loaded[1].ID, "LF-002")
	}
}

func TestLoadFeedback_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	feedback, err := LoadFeedback(dir)
	if err != nil {
		t.Fatalf("LoadFeedback on empty dir failed: %v", err)
	}
	if len(feedback) != 0 {
		t.Errorf("expected 0 feedback from empty dir, got %d", len(feedback))
	}
}
