package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/artifacts"
)

// newTestOrchestrator creates an Orchestrator with temp directories
// and a fixed clock for deterministic testing.
func newTestOrchestrator(t *testing.T, agentFiles []string, binaries map[string]bool) *Orchestrator {
	t.Helper()

	base := t.TempDir()
	wfDir := filepath.Join(base, "workflows")
	artDir := filepath.Join(base, "artifacts")
	agentDir := filepath.Join(base, "agents")

	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("create agent dir: %v", err)
	}

	for _, name := range agentFiles {
		if err := os.WriteFile(filepath.Join(agentDir, name), []byte("# agent"), 0644); err != nil {
			t.Fatalf("create agent %s: %v", name, err)
		}
	}

	fixedTime := time.Date(2026, 3, 20, 14, 30, 0, 0, time.UTC)
	callCount := 0

	return &Orchestrator{
		WorkflowDir: wfDir,
		ArtifactDir: artDir,
		AgentDir:    agentDir,
		Now: func() time.Time {
			callCount++
			return fixedTime.Add(time.Duration(callCount) * time.Second)
		},
		Stdout: os.Stdout,
		LookPath: func(name string) (string, error) {
			if binaries[name] {
				return "/usr/local/bin/" + name, nil
			}
			return "", fmt.Errorf("not found: %s", name)
		},
	}
}

// allAgentFiles returns the agent files needed for all heroes.
var allAgentFiles = []string{
	"muti-mind-po.md",
	"cobalt-crush-dev.md",
	"divisor-guard.md",
	"mx-f-coach.md",
}

// allBinaries returns the binaries needed for all heroes.
var allBinaries = map[string]bool{"gaze": true, "mxf": true}

func TestOrchestrator_Start_AllHeroes(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/health-check", "BI-042")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wf := result.Workflow
	if wf.Status != StatusActive {
		t.Errorf("Status = %q, want %q", wf.Status, StatusActive)
	}
	if len(wf.Stages) != 6 {
		t.Fatalf("expected 6 stages, got %d", len(wf.Stages))
	}

	// No stages should be skipped when all heroes are present
	for _, stage := range wf.Stages {
		if stage.Status == StatusSkipped {
			t.Errorf("stage %q should not be skipped", stage.StageName)
		}
	}

	// First stage should be active
	if wf.Stages[0].Status != StatusActive {
		t.Errorf("first stage status = %q, want %q", wf.Stages[0].Status, StatusActive)
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", result.Warnings)
	}

	if wf.FeatureBranch != "feat/health-check" {
		t.Errorf("FeatureBranch = %q, want %q", wf.FeatureBranch, "feat/health-check")
	}
	if wf.BacklogItemID != "BI-042" {
		t.Errorf("BacklogItemID = %q, want %q", wf.BacklogItemID, "BI-042")
	}
}

func TestOrchestrator_Start_MissingHeroes(t *testing.T) {
	// Only muti-mind and cobalt-crush available
	orch := newTestOrchestrator(t, []string{"muti-mind-po.md", "cobalt-crush-dev.md"}, map[string]bool{})

	result, err := orch.Start("feat/test", "BI-001")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wf := result.Workflow

	// Gaze (validate), Divisor (review), Mx F (measure) should be skipped
	skippedStages := make(map[string]bool)
	for _, stage := range wf.Stages {
		if stage.Status == StatusSkipped {
			skippedStages[stage.StageName] = true
			if stage.SkipReason == "" {
				t.Errorf("skipped stage %q has no skip_reason", stage.StageName)
			}
		}
	}

	if !skippedStages[StageValidate] {
		t.Error("validate stage should be skipped (gaze unavailable)")
	}
	if !skippedStages[StageReview] {
		t.Error("review stage should be skipped (divisor unavailable)")
	}
	if !skippedStages[StageReflect] {
		t.Error("reflect stage should be skipped (mx-f unavailable)")
	}

	if len(result.Warnings) < 3 {
		t.Errorf("expected at least 3 warnings for skipped stages, got %d", len(result.Warnings))
	}
}

func TestOrchestrator_Advance_ThroughAllStages(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/full", "BI-010")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// With execution modes, the workflow pauses at the review→accept
	// boundary (swarm→human). The advance sequence is:
	//   1: define(human) → implement(swarm)
	//   2: implement(swarm) → validate(swarm)
	//   3: validate(swarm) → review(swarm)
	//   4: review(swarm) → awaiting_human (checkpoint before accept)
	//   5: resume → accept(human) activates
	//   6: accept(human) → reflect(swarm)
	//   7: reflect(swarm) → completed
	awaitingHumanCount := 0
	for i := 0; i < 7; i++ {
		result, err = orch.Advance(wfID)
		if err != nil {
			t.Fatalf("Advance %d failed: %v", i+1, err)
		}
		if result.Workflow.Status == StatusAwaitingHuman {
			awaitingHumanCount++
		}
	}

	if awaitingHumanCount != 1 {
		t.Errorf("expected workflow to pass through awaiting_human exactly once, got %d", awaitingHumanCount)
	}

	// After 7 advances, workflow should be completed
	wf := result.Workflow
	if wf.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", wf.Status, StatusCompleted)
	}

	// All stages should be completed
	for _, stage := range wf.Stages {
		if stage.Status != StatusCompleted {
			t.Errorf("stage %q status = %q, want %q", stage.StageName, stage.Status, StatusCompleted)
		}
	}

	// Final stage should be "reflect" (not "measure")
	lastStage := wf.Stages[len(wf.Stages)-1]
	if lastStage.StageName != StageReflect {
		t.Errorf("last stage = %q, want %q", lastStage.StageName, StageReflect)
	}
}

func TestOrchestrator_Advance_SkipsUnavailable(t *testing.T) {
	// Only muti-mind and cobalt-crush available
	orch := newTestOrchestrator(t, []string{"muti-mind-po.md", "cobalt-crush-dev.md"}, map[string]bool{})

	result, err := orch.Start("feat/partial", "BI-020")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Advance: define → implement (skip validate, review, reflect) → accept → complete
	// Stage 0 (define) is active, advance completes it and moves to stage 1 (implement)
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 1 failed: %v", err)
	}
	if result.Workflow.Stages[1].Status != StatusActive {
		t.Errorf("after advance 1, stage 1 (implement) should be active, got %q", result.Workflow.Stages[1].Status)
	}

	// Advance: implement (swarm) → accept (human) — triggers checkpoint
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 2 failed: %v", err)
	}
	if result.Workflow.Status != StatusAwaitingHuman {
		t.Errorf("after advance 2, status = %q, want %q", result.Workflow.Status, StatusAwaitingHuman)
	}

	// Advance: resume from checkpoint → accept activates
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 3 (resume) failed: %v", err)
	}
	if result.Workflow.Stages[4].Status != StatusActive {
		t.Errorf("after advance 3, stage 4 (accept) should be active, got %q", result.Workflow.Stages[4].Status)
	}

	// Advance: accept → complete (reflect skipped)
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 4 failed: %v", err)
	}

	if result.Workflow.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", result.Workflow.Status, StatusCompleted)
	}
}

func TestOrchestrator_Escalate(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/escalate", "BI-030")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	if err := orch.Escalate(wfID, "manual review needed"); err != nil {
		t.Fatalf("Escalate failed: %v", err)
	}

	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if wf.Status != StatusEscalated {
		t.Errorf("Status = %q, want %q", wf.Status, StatusEscalated)
	}
	if wf.CompletedAt == nil {
		t.Error("CompletedAt should be set on escalation")
	}
}

func TestOrchestrator_Complete_ProducesRecord(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/record", "BI-040")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Advance through all stages (7 advances with checkpoint)
	for i := 0; i < 7; i++ {
		_, err = orch.Advance(wfID)
		if err != nil {
			t.Fatalf("Advance %d failed: %v", i+1, err)
		}
	}

	// Verify workflow-record artifact was written
	artDir := orch.ArtifactDir
	paths, err := findJSONFiles(artDir)
	if err != nil {
		t.Fatalf("find artifacts: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("expected at least one workflow-record artifact")
	}

	// Read and validate the artifact envelope
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}

	var env artifacts.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	if env.Hero == "" {
		t.Error("envelope Hero should not be empty")
	}
	if env.ArtifactType != "workflow-record" {
		t.Errorf("ArtifactType = %q, want %q", env.ArtifactType, "workflow-record")
	}

	var record WorkflowRecord
	if err := json.Unmarshal(env.Payload, &record); err != nil {
		t.Fatalf("unmarshal payload into WorkflowRecord: %v", err)
	}
	if record.Outcome != "shipped" {
		t.Errorf("record.Outcome = %q, want %q", record.Outcome, "shipped")
	}
}

func TestOrchestrator_Start_NoHeroes(t *testing.T) {
	orch := newTestOrchestrator(t, nil, map[string]bool{})

	result, err := orch.Start("feat/empty", "BI-050")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// All stages should be skipped
	for _, stage := range result.Workflow.Stages {
		if stage.Status != StatusSkipped {
			t.Errorf("stage %q should be skipped, got %q", stage.StageName, stage.Status)
		}
	}

	// Should have a warning about no heroes
	foundNoHeroWarning := false
	for _, w := range result.Warnings {
		if w == "no heroes available — all stages skipped" {
			foundNoHeroWarning = true
			break
		}
	}
	if !foundNoHeroWarning {
		t.Errorf("expected 'no heroes available' warning, got %v", result.Warnings)
	}
}

func TestGenerateWorkflowRecord_CompletedWorkflow(t *testing.T) {
	start := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 20, 16, 30, 0, 0, time.UTC)

	wf := &WorkflowInstance{
		WorkflowID:    "wf-completed-001",
		BacklogItemID: "BI-042",
		StartedAt:     start,
		Status:        StatusCompleted,
		Stages: []WorkflowStage{
			{StageName: StageDefine, Hero: "muti-mind", Status: StatusCompleted,
				ArtifactsProduced: []string{"spec.md"}},
			{StageName: StageImplement, Hero: "cobalt-crush", Status: StatusCompleted,
				ArtifactsProduced: []string{"main.go"}},
			{StageName: StageValidate, Hero: "gaze", Status: StatusCompleted,
				ArtifactsProduced: []string{"quality-report.json"}},
			{StageName: StageReview, Hero: "divisor", Status: StatusCompleted},
			{StageName: StageAccept, Hero: "muti-mind", Status: StatusCompleted},
			{StageName: StageReflect, Hero: "mx-f", Status: StatusCompleted},
		},
	}

	record := GenerateWorkflowRecord(wf, end)

	if record.WorkflowID != "wf-completed-001" {
		t.Errorf("WorkflowID = %q, want %q", record.WorkflowID, "wf-completed-001")
	}
	if record.BacklogItemID != "BI-042" {
		t.Errorf("BacklogItemID = %q, want %q", record.BacklogItemID, "BI-042")
	}
	if record.Outcome != OutcomeShipped {
		t.Errorf("Outcome = %q, want %q", record.Outcome, OutcomeShipped)
	}
	if len(record.Artifacts) != 3 {
		t.Errorf("expected 3 artifacts, got %d", len(record.Artifacts))
	}
	if record.TotalElapsedTime == "" {
		t.Error("TotalElapsedTime should not be empty")
	}
}

func TestGenerateWorkflowRecord_EscalatedWorkflow(t *testing.T) {
	start := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC)

	wf := &WorkflowInstance{
		WorkflowID: "wf-escalated-001",
		StartedAt:  start,
		Status:     StatusEscalated,
		Stages: []WorkflowStage{
			{StageName: StageDefine, Hero: "muti-mind", Status: StatusCompleted},
			{StageName: StageImplement, Hero: "cobalt-crush", Status: StatusCompleted},
			{StageName: StageValidate, Hero: "gaze", Status: StatusFailed,
				Error: "escalated: max iterations"},
		},
	}

	record := GenerateWorkflowRecord(wf, end)

	if record.Outcome != OutcomeAbandoned {
		t.Errorf("Outcome = %q, want %q", record.Outcome, OutcomeAbandoned)
	}
}

func TestGenerateWorkflowRecord_RejectedWorkflow(t *testing.T) {
	start := time.Date(2026, 3, 20, 14, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 20, 16, 0, 0, 0, time.UTC)

	wf := &WorkflowInstance{
		WorkflowID: "wf-rejected-001",
		StartedAt:  start,
		Status:     StatusCompleted,
		Stages: []WorkflowStage{
			{StageName: StageDefine, Hero: "muti-mind", Status: StatusCompleted},
			{StageName: StageImplement, Hero: "cobalt-crush", Status: StatusCompleted},
			{StageName: StageValidate, Hero: "gaze", Status: StatusCompleted},
			{StageName: StageReview, Hero: "divisor", Status: StatusCompleted},
			{StageName: StageAccept, Hero: "muti-mind", Status: StatusFailed},
			{StageName: StageReflect, Hero: "mx-f", Status: StatusSkipped},
		},
	}

	record := GenerateWorkflowRecord(wf, end)

	if record.Outcome != OutcomeRejected {
		t.Errorf("Outcome = %q, want %q", record.Outcome, OutcomeRejected)
	}
}

func TestOrchestrator_ConcurrentWorkflows_Isolated(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	// Start two workflows on different branches
	result1, err := orch.Start("feat/alpha", "BI-100")
	if err != nil {
		t.Fatalf("Start alpha failed: %v", err)
	}
	result2, err := orch.Start("feat/beta", "BI-200")
	if err != nil {
		t.Fatalf("Start beta failed: %v", err)
	}

	wfID1 := result1.Workflow.WorkflowID
	wfID2 := result2.Workflow.WorkflowID

	if wfID1 == wfID2 {
		t.Error("workflow IDs should be unique")
	}

	// Advance alpha once
	_, err = orch.Advance(wfID1)
	if err != nil {
		t.Fatalf("Advance alpha failed: %v", err)
	}

	// Verify alpha advanced but beta didn't
	wf1, err := orch.Status(wfID1)
	if err != nil {
		t.Fatalf("Status alpha failed: %v", err)
	}
	wf2, err := orch.Status(wfID2)
	if err != nil {
		t.Fatalf("Status beta failed: %v", err)
	}

	if wf1.CurrentStage == wf2.CurrentStage && wf1.Stages[0].Status == wf2.Stages[0].Status {
		t.Error("workflows should have different states after advancing only one")
	}

	// Verify Latest returns correct workflow per branch
	latest1, err := orch.store().Latest("feat/alpha")
	if err != nil {
		t.Fatalf("Latest alpha failed: %v", err)
	}
	if latest1 == nil || latest1.WorkflowID != wfID1 {
		t.Errorf("Latest(feat/alpha) = %v, want %q", latest1, wfID1)
	}

	latest2, err := orch.store().Latest("feat/beta")
	if err != nil {
		t.Fatalf("Latest beta failed: %v", err)
	}
	if latest2 == nil || latest2.WorkflowID != wfID2 {
		t.Errorf("Latest(feat/beta) = %v, want %q", latest2, wfID2)
	}
}

func TestOrchestrator_Advance_MaxIterations_Escalates(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/max-iter", "BI-060")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Manually set iteration_count to MaxIterations to trigger escalation
	wf, err := orch.store().Load(wfID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	wf.IterationCount = MaxIterations

	// Advance to the review stage (stages 0-2 need to be completed first)
	// Set stages 0-2 as completed, stage 3 (review) as pending
	now := time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		wf.Stages[i].Status = StatusCompleted
		wf.Stages[i].CompletedAt = &now
	}
	wf.Stages[3].Status = StatusPending // review
	wf.CurrentStage = 2                 // currently at validate (completed)

	if err := orch.store().Save(wf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Advance should trigger escalation when moving to review with max iterations
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance failed: %v", err)
	}

	if result.Workflow.Status != StatusEscalated {
		t.Errorf("Status = %q, want %q", result.Workflow.Status, StatusEscalated)
	}

	foundEscalationWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "escalat") {
			foundEscalationWarning = true
			break
		}
	}
	if !foundEscalationWarning {
		t.Error("expected escalation warning containing 'escalat'")
	}
}

func TestOrchestrator_HandleAcceptanceRejection(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/reject", "BI-070")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	decision := Decision{
		Type:      "acceptance-decision",
		Hero:      "muti-mind",
		Result:    "reject",
		Rationale: "missing edge case handling for empty input",
		Iteration: 1,
		Timestamp: time.Now().UTC(),
	}

	if err := orch.HandleAcceptanceRejection(wfID, decision); err != nil {
		t.Fatalf("HandleAcceptanceRejection failed: %v", err)
	}

	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if wf.Status != StatusFailed {
		t.Errorf("Status = %q, want %q", wf.Status, StatusFailed)
	}

	// Verify accept stage has the rejection reason
	for _, stage := range wf.Stages {
		if stage.StageName == StageAccept {
			if stage.Status != StatusFailed {
				t.Errorf("accept stage status = %q, want %q", stage.Status, StatusFailed)
			}
			if !strings.Contains(stage.Error, decision.Rationale) {
				t.Errorf("accept stage error = %q, want it to contain rationale %q", stage.Error, decision.Rationale)
			}
			break
		}
	}
}

func TestOrchestrator_HandleContradiction(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/conflict", "BI-080")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	conflict := "Muti-Mind says 'ship quickly' but Gaze says 'quality is insufficient' — CRAP score 45 on critical path"

	if err := orch.HandleContradiction(wfID, conflict); err != nil {
		t.Fatalf("HandleContradiction failed: %v", err)
	}

	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if wf.Status != StatusEscalated {
		t.Errorf("Status = %q, want %q", wf.Status, StatusEscalated)
	}

	// Verify the contradiction is recorded in the current stage
	stage := wf.Stages[wf.CurrentStage]
	if stage.Status != StatusFailed {
		t.Errorf("current stage status = %q, want %q", stage.Status, StatusFailed)
	}
	if !strings.Contains(stage.Error, conflict) {
		t.Errorf("current stage error = %q, want it to contain conflict %q", stage.Error, conflict)
	}
}

func TestOrchestrator_Skip_ValidStage(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/skip", "BI-090")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Skip stage 2 (validate)
	if err := orch.Skip(wfID, 2, "not needed for this feature"); err != nil {
		t.Fatalf("Skip failed: %v", err)
	}

	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if wf.Stages[2].Status != StatusSkipped {
		t.Errorf("stage 2 status = %q, want %q", wf.Stages[2].Status, StatusSkipped)
	}
	if wf.Stages[2].SkipReason != "not needed for this feature" {
		t.Errorf("stage 2 skip_reason = %q, want %q", wf.Stages[2].SkipReason, "not needed for this feature")
	}
}

func TestOrchestrator_Advance_NonActiveWorkflow(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/done", "BI-091")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Advance through all stages to complete the workflow (7 with checkpoint)
	for i := 0; i < 7; i++ {
		_, err = orch.Advance(wfID)
		if err != nil {
			t.Fatalf("Advance %d failed: %v", i+1, err)
		}
	}

	// Verify workflow is completed
	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if wf.Status != StatusCompleted {
		t.Fatalf("Status = %q, want %q", wf.Status, StatusCompleted)
	}

	// Advance on a completed workflow should return an error
	_, err = orch.Advance(wfID)
	if err == nil {
		t.Fatal("Advance on completed workflow should return an error")
	}
	if !strings.Contains(err.Error(), "not active") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "not active")
	}
}

func TestSanitizeBranch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"feat/login", "feat-login"},
		{"fix bug", "fix-bug"},
		{"", ""},
		{"clean", "clean"},
		{"feat/~special^chars", "feat--special-chars"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeBranch(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeBranch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- User Story 1: Swarm Delegation After Clarify ---

func TestOrchestrator_NewWorkflow_SetsExecutionModes(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	wf := orch.NewWorkflow("feat/modes", "BI-100")

	expectedModes := StageExecutionModeMap()
	for _, stage := range wf.Stages {
		want := expectedModes[stage.StageName]
		if stage.ExecutionMode != want {
			t.Errorf("stage %q ExecutionMode = %q, want %q", stage.StageName, stage.ExecutionMode, want)
		}
	}
}

func TestOrchestrator_Advance_PausesAtHumanCheckpoint(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/checkpoint", "BI-101")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Advance 1: complete define (human) → implement (swarm) activates
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 1 (define→implement) failed: %v", err)
	}
	if result.Workflow.Status != StatusActive {
		t.Errorf("after advance 1, status = %q, want %q", result.Workflow.Status, StatusActive)
	}

	// Advance 2: complete implement (swarm) → validate (swarm) activates
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 2 (implement→validate) failed: %v", err)
	}
	if result.Workflow.Status != StatusActive {
		t.Errorf("after advance 2, status = %q, want %q", result.Workflow.Status, StatusActive)
	}

	// Advance 3: complete validate (swarm) → review (swarm) activates
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 3 (validate→review) failed: %v", err)
	}
	if result.Workflow.Status != StatusActive {
		t.Errorf("after advance 3, status = %q, want %q", result.Workflow.Status, StatusActive)
	}

	// Advance 4: complete review (swarm) → next is accept (human)
	// This should trigger awaiting_human checkpoint
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 4 (review→checkpoint) failed: %v", err)
	}
	if result.Workflow.Status != StatusAwaitingHuman {
		t.Errorf("after advance 4, status = %q, want %q", result.Workflow.Status, StatusAwaitingHuman)
	}

	// Accept stage should still be pending (not activated)
	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if wf.Stages[4].Status != StatusPending {
		t.Errorf("accept stage status = %q, want %q", wf.Stages[4].Status, StatusPending)
	}
}

func TestOrchestrator_Advance_ResumesFromCheckpoint(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/resume", "BI-102")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Advance through define, implement, validate, review to reach checkpoint
	for i := 0; i < 4; i++ {
		result, err = orch.Advance(wfID)
		if err != nil {
			t.Fatalf("Advance %d failed: %v", i+1, err)
		}
	}

	// Verify we're at the checkpoint
	if result.Workflow.Status != StatusAwaitingHuman {
		t.Fatalf("expected awaiting_human, got %q", result.Workflow.Status)
	}

	// Advance from checkpoint — should resume and activate accept
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance from checkpoint failed: %v", err)
	}

	if result.Workflow.Status != StatusActive {
		t.Errorf("after resume, status = %q, want %q", result.Workflow.Status, StatusActive)
	}

	// Accept stage should now be active
	wf, err := orch.Status(wfID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if wf.Stages[4].Status != StatusActive {
		t.Errorf("accept stage status = %q, want %q", wf.Stages[4].Status, StatusActive)
	}
}

func TestOrchestrator_Advance_SwarmToSwarmNoPause(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/swarm-flow", "BI-103")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Advance 1: define (human) → implement (swarm) — no checkpoint
	// (human→swarm does not trigger checkpoint)
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 1 failed: %v", err)
	}
	if result.Workflow.Status == StatusAwaitingHuman {
		t.Error("human→swarm transition should NOT trigger awaiting_human")
	}

	// Advance 2: implement (swarm) → validate (swarm) — no checkpoint
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 2 failed: %v", err)
	}
	if result.Workflow.Status == StatusAwaitingHuman {
		t.Error("swarm→swarm (implement→validate) should NOT trigger awaiting_human")
	}

	// Advance 3: validate (swarm) → review (swarm) — no checkpoint
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 3 failed: %v", err)
	}
	if result.Workflow.Status == StatusAwaitingHuman {
		t.Error("swarm→swarm (validate→review) should NOT trigger awaiting_human")
	}
}

// --- User Story 2: Execution Mode Per Stage ---

func TestOrchestrator_Advance_LegacyWorkflowNoCheckpoints(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/legacy", "BI-200")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Simulate a legacy workflow by clearing all execution_mode fields.
	// Per FR-010: empty execution mode is treated as "human" for backward
	// compatibility, so no checkpoint pausing should occur.
	wf, err := orch.store().Load(wfID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	for i := range wf.Stages {
		wf.Stages[i].ExecutionMode = ""
	}
	if err := orch.store().Save(wf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Advance through all 6 stages — no awaiting_human should occur
	// because all stages are effectively human-mode (empty = human).
	for i := 0; i < 6; i++ {
		result, err = orch.Advance(wfID)
		if err != nil {
			t.Fatalf("Advance %d failed: %v", i+1, err)
		}
		if result.Workflow.Status == StatusAwaitingHuman {
			t.Fatalf("legacy workflow should NOT trigger awaiting_human at advance %d", i+1)
		}
	}

	if result.Workflow.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", result.Workflow.Status, StatusCompleted)
	}
}

func TestOrchestrator_Advance_AllSwarmSkipped_NoCheckpoint(t *testing.T) {
	// Only muti-mind available — all swarm-mode heroes are unavailable.
	// Per FR-014: when all swarm-mode stages between two human checkpoints
	// are skipped, the workflow transitions directly to the next human-mode
	// stage without entering awaiting_human.
	orch := newTestOrchestrator(t, []string{"muti-mind-po.md"}, map[string]bool{})

	result, err := orch.Start("feat/no-swarm", "BI-201")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Only define and accept are available (both human-mode).
	// implement, validate, review, reflect are all skipped.
	// Advance from define (human) → accept (human) should NOT trigger
	// awaiting_human because the completed stage is human-mode.
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 1 failed: %v", err)
	}
	if result.Workflow.Status == StatusAwaitingHuman {
		t.Error("human→human transition (all swarm skipped) should NOT trigger awaiting_human")
	}
	if result.Workflow.Stages[4].Status != StatusActive {
		t.Errorf("accept stage status = %q, want %q", result.Workflow.Stages[4].Status, StatusActive)
	}

	// Advance from accept → complete (reflect skipped)
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance 2 failed: %v", err)
	}
	if result.Workflow.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", result.Workflow.Status, StatusCompleted)
	}
}

func TestOrchestrator_Advance_EscalationWithExecutionModes(t *testing.T) {
	orch := newTestOrchestrator(t, allAgentFiles, allBinaries)

	result, err := orch.Start("feat/esc-modes", "BI-202")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	wfID := result.Workflow.WorkflowID

	// Manually set up: stages 0-2 completed, stage 3 (review) pending,
	// iteration count at max. Verify escalation fires regardless of
	// execution mode.
	wf, err := orch.store().Load(wfID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	wf.IterationCount = MaxIterations
	now := time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		wf.Stages[i].Status = StatusCompleted
		wf.Stages[i].CompletedAt = &now
	}
	wf.Stages[3].Status = StatusPending // review
	wf.CurrentStage = 2                 // currently at validate (completed)
	if err := orch.store().Save(wf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Advance should trigger escalation at review regardless of execution mode
	result, err = orch.Advance(wfID)
	if err != nil {
		t.Fatalf("Advance failed: %v", err)
	}

	if result.Workflow.Status != StatusEscalated {
		t.Errorf("Status = %q, want %q", result.Workflow.Status, StatusEscalated)
	}
}

// findJSONFiles walks a directory and returns all .json file paths.
func findJSONFiles(dir string) ([]string, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
