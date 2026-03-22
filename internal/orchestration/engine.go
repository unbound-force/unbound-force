package orchestration

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/unbound-force/unbound-force/internal/artifacts"
	"github.com/unbound-force/unbound-force/internal/sync"
)

// MaxIterations is the maximum number of review-fix iterations
// before the workflow is escalated to manual review.
const MaxIterations = 3

// WorkflowResult is returned by Start and Advance to report
// the outcome of the operation.
type WorkflowResult struct {
	Workflow  *WorkflowInstance `json:"workflow"`
	StagesRun int               `json:"stages_run"`
	Warnings  []string          `json:"warnings,omitempty"`
}

// Orchestrator manages the hero lifecycle workflow. It coordinates
// stage transitions, hero detection, artifact tracking, and
// workflow state persistence.
//
// Design decision: All dependencies are injected via struct fields
// (DI per SOLID Dependency Inversion Principle). This enables
// testing with t.TempDir() paths and stubbed clocks.
type Orchestrator struct {
	WorkflowDir string        // .unbound-force/workflows/
	ArtifactDir string        // .unbound-force/artifacts/
	AgentDir    string        // .opencode/agents/
	GHRunner    sync.GHRunner // GitHub CLI interface (unused in v1.0.0)
	Now         func() time.Time
	Stdout      io.Writer
	LookPath    func(string) (string, error) // defaults to exec.LookPath
}

// store returns a WorkflowStore for the orchestrator's workflow directory.
func (o *Orchestrator) store() *WorkflowStore {
	return &WorkflowStore{Dir: o.WorkflowDir}
}

// now returns the current time, using the injected clock if available.
func (o *Orchestrator) now() time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return time.Now().UTC()
}

// lookPath returns the injected lookPath function or exec.LookPath.
func (o *Orchestrator) lookPath() func(string) (string, error) {
	if o.LookPath != nil {
		return o.LookPath
	}
	return exec.LookPath
}

// NewWorkflow creates a new WorkflowInstance with 6 stages,
// detects available heroes, and marks unavailable stages as skipped.
func (o *Orchestrator) NewWorkflow(branch, backlogItemID string) *WorkflowInstance {
	now := o.now()
	id := fmt.Sprintf("wf-%s-%s", sanitizeBranch(branch), now.Format("20060102T150405"))

	heroes, _ := DetectHeroes(o.AgentDir, o.lookPath())
	heroAvail := make(map[string]bool)
	var availableNames []string
	for _, h := range heroes {
		heroAvail[h.Name] = h.Available
		if h.Available {
			availableNames = append(availableNames, h.Name)
		}
	}

	stageHeroes := StageHeroMap()
	stageModes := StageExecutionModeMap()
	order := StageOrder()
	stages := make([]WorkflowStage, len(order))
	for i, stageName := range order {
		hero := stageHeroes[stageName]
		stage := WorkflowStage{
			StageName:     stageName,
			Hero:          hero,
			Status:        StatusPending,
			ExecutionMode: stageModes[stageName],
		}

		if !heroAvail[hero] {
			stage.Status = StatusSkipped
			stage.SkipReason = fmt.Sprintf("hero %q unavailable", hero)
		}

		stages[i] = stage
	}

	return &WorkflowInstance{
		WorkflowID:      id,
		FeatureBranch:   branch,
		BacklogItemID:   backlogItemID,
		Stages:          stages,
		CurrentStage:    0,
		StartedAt:       now,
		Status:          StatusActive,
		AvailableHeroes: availableNames,
		IterationCount:  0,
	}
}

// Start creates a new workflow, activates the first non-skipped stage,
// saves it, and returns the result with any warnings.
func (o *Orchestrator) Start(branch, backlogItemID string) (*WorkflowResult, error) {
	wf := o.NewWorkflow(branch, backlogItemID)

	var warnings []string
	for _, stage := range wf.Stages {
		if stage.Status == StatusSkipped {
			warnings = append(warnings, fmt.Sprintf("stage %q skipped: %s", stage.StageName, stage.SkipReason))
		}
	}

	// Activate the first non-skipped stage
	activated := false
	for i := range wf.Stages {
		if wf.Stages[i].Status == StatusPending {
			now := o.now()
			wf.Stages[i].Status = StatusActive
			wf.Stages[i].StartedAt = &now
			wf.CurrentStage = i
			activated = true
			break
		}
	}

	if !activated {
		warnings = append(warnings, "no heroes available — all stages skipped")
	}

	if err := o.store().Save(wf); err != nil {
		return nil, fmt.Errorf("save workflow: %w", err)
	}

	return &WorkflowResult{
		Workflow:  wf,
		StagesRun: 0,
		Warnings:  warnings,
	}, nil
}

// Advance completes the current stage and moves to the next
// non-skipped stage. If the current stage is the last, it calls
// Complete. If the review stage has reached max iterations, it
// escalates.
func (o *Orchestrator) Advance(workflowID string) (*WorkflowResult, error) {
	wf, err := o.store().Load(workflowID)
	if err != nil {
		return nil, fmt.Errorf("load workflow: %w", err)
	}

	// Accept both active and awaiting_human workflows. Reject completed,
	// failed, and escalated workflows. Per FR-005 and research.md R1:
	// Advance() serves as both the normal transition and the resume API.
	if wf.Status != StatusActive && wf.Status != StatusAwaitingHuman {
		return nil, fmt.Errorf("workflow %q is not active (status: %s)", workflowID, wf.Status)
	}

	// Resume path: when the workflow is paused at a human checkpoint,
	// skip the current-stage-completion logic and proceed directly to
	// finding and activating the next pending human-mode stage.
	resuming := wf.Status == StatusAwaitingHuman

	// Complete the current stage (skip when resuming from checkpoint)
	current := wf.CurrentStage
	if !resuming {
		if current >= 0 && current < len(wf.Stages) && wf.Stages[current].Status == StatusActive {
			now := o.now()
			wf.Stages[current].Status = StatusCompleted
			wf.Stages[current].CompletedAt = &now

			// Record artifacts consumed from previous stage's hero.
			// This is metadata tracking — not actual hero invocation.
			if current > 0 {
				prevStage := wf.Stages[current-1]
				if len(prevStage.ArtifactsProduced) > 0 {
					wf.Stages[current].ArtifactsConsumed = append(
						wf.Stages[current].ArtifactsConsumed,
						prevStage.ArtifactsProduced...,
					)
				}
			}
		}
	}

	// Find the next non-skipped stage
	stagesRun := 1
	nextFound := false
	for i := current + 1; i < len(wf.Stages); i++ {
		if wf.Stages[i].Status == StatusSkipped {
			continue
		}
		if wf.Stages[i].Status == StatusPending {
			// Checkpoint detection: when completing a swarm-mode stage
			// and the next non-skipped stage is human-mode, pause the
			// workflow at the boundary instead of activating the next
			// stage. Per FR-004 and research.md R1.
			// Skip checkpoint detection when resuming from an existing
			// checkpoint — the human has explicitly chosen to advance.
			completedMode := effectiveMode(wf.Stages[current].ExecutionMode)
			nextMode := effectiveMode(wf.Stages[i].ExecutionMode)
			if !resuming && completedMode == ModeSwarm && nextMode == ModeHuman {
				wf.Status = StatusAwaitingHuman
				nextFound = true
				if err := o.store().Save(wf); err != nil {
					return nil, fmt.Errorf("save workflow at checkpoint: %w", err)
				}
				return &WorkflowResult{
					Workflow:  wf,
					StagesRun: stagesRun,
				}, nil
			}

			now := o.now()
			wf.Stages[i].Status = StatusActive
			wf.Stages[i].StartedAt = &now
			wf.CurrentStage = i
			nextFound = true

			// When resuming from a checkpoint, restore active status.
			if resuming {
				wf.Status = StatusActive
			}

			// Track review iterations
			if wf.Stages[i].StageName == StageReview {
				wf.IterationCount++
				if wf.IterationCount > MaxIterations {
					return o.doEscalate(wf, "maximum review iterations reached")
				}
			}
			break
		}
	}

	var warnings []string
	if !nextFound {
		// Save the workflow with the completed current stage first,
		// then finalize. Complete reloads from disk, so the stage
		// completion must be persisted before calling it.
		if err := o.store().Save(wf); err != nil {
			return nil, fmt.Errorf("save workflow before complete: %w", err)
		}

		_, err := o.Complete(workflowID)
		if err != nil {
			return nil, fmt.Errorf("complete workflow: %w", err)
		}

		// Reload to get the completed state
		wf, err = o.store().Load(workflowID)
		if err != nil {
			return nil, fmt.Errorf("reload workflow after complete: %w", err)
		}
	} else {
		if err := o.store().Save(wf); err != nil {
			return nil, fmt.Errorf("save workflow: %w", err)
		}
	}

	return &WorkflowResult{
		Workflow:  wf,
		StagesRun: stagesRun,
		Warnings:  warnings,
	}, nil
}

// Skip marks a stage as skipped with the given reason.
func (o *Orchestrator) Skip(workflowID string, stage int, reason string) error {
	wf, err := o.store().Load(workflowID)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	if stage < 0 || stage >= len(wf.Stages) {
		return fmt.Errorf("stage index %d out of range [0, %d)", stage, len(wf.Stages))
	}

	wf.Stages[stage].Status = StatusSkipped
	wf.Stages[stage].SkipReason = reason

	return o.store().Save(wf)
}

// Escalate sets the workflow status to escalated with the given reason.
func (o *Orchestrator) Escalate(workflowID, reason string) error {
	wf, err := o.store().Load(workflowID)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	now := o.now()
	wf.Status = StatusEscalated
	wf.CompletedAt = &now

	// Record the escalation in the current stage's error field
	if wf.CurrentStage >= 0 && wf.CurrentStage < len(wf.Stages) {
		wf.Stages[wf.CurrentStage].Status = StatusFailed
		wf.Stages[wf.CurrentStage].Error = fmt.Sprintf("escalated: %s", reason)
		wf.Stages[wf.CurrentStage].CompletedAt = &now
	}

	return o.store().Save(wf)
}

// Complete finalizes a workflow, produces a workflow-record artifact
// via WriteArtifact, and returns the record.
func (o *Orchestrator) Complete(workflowID string) (*WorkflowRecord, error) {
	wf, err := o.store().Load(workflowID)
	if err != nil {
		return nil, fmt.Errorf("load workflow: %w", err)
	}

	now := o.now()
	wf.Status = StatusCompleted
	wf.CompletedAt = &now

	record := GenerateWorkflowRecord(wf, now)

	// Write workflow-record artifact with context
	ctx := &artifacts.ArtifactContext{
		Branch:        wf.FeatureBranch,
		BacklogItemID: wf.BacklogItemID,
		WorkflowID:    wf.WorkflowID,
	}
	if err := artifacts.WriteArtifactWithContext(
		o.ArtifactDir, "orchestration", "workflow-record",
		wf.WorkflowID, record, ctx,
	); err != nil {
		return nil, fmt.Errorf("write workflow-record artifact: %w", err)
	}

	if err := o.store().Save(wf); err != nil {
		return nil, fmt.Errorf("save completed workflow: %w", err)
	}

	return record, nil
}

// Status loads and returns the current state of a workflow.
func (o *Orchestrator) Status(workflowID string) (*WorkflowInstance, error) {
	return o.store().Load(workflowID)
}

// List returns all workflows, optionally filtered by status.
func (o *Orchestrator) List(statusFilter string) ([]WorkflowInstance, error) {
	return o.store().List(statusFilter)
}

// doEscalate is an internal helper that escalates a workflow and
// returns a WorkflowResult. Used by Advance when max iterations
// are reached.
func (o *Orchestrator) doEscalate(wf *WorkflowInstance, reason string) (*WorkflowResult, error) {
	now := o.now()
	wf.Status = StatusEscalated
	wf.CompletedAt = &now

	if wf.CurrentStage >= 0 && wf.CurrentStage < len(wf.Stages) {
		wf.Stages[wf.CurrentStage].Status = StatusFailed
		wf.Stages[wf.CurrentStage].Error = fmt.Sprintf("escalated: %s", reason)
		wf.Stages[wf.CurrentStage].CompletedAt = &now
	}

	if err := o.store().Save(wf); err != nil {
		return nil, fmt.Errorf("save escalated workflow: %w", err)
	}

	return &WorkflowResult{
		Workflow:  wf,
		StagesRun: 0,
		Warnings:  []string{fmt.Sprintf("workflow escalated: %s", reason)},
	}, nil
}

// HandleHeroUnavailable marks a stage as skipped due to hero
// unavailability. This handles mid-workflow detection (e.g., a
// hero binary was uninstalled after workflow start).
func (o *Orchestrator) HandleHeroUnavailable(workflowID string, stageIdx int) error {
	return o.Skip(workflowID, stageIdx, "hero unavailable")
}

// HandleMaxIterations escalates a workflow when the review-implement
// loop has exceeded MaxIterations. Includes a summary of unresolved
// findings in the escalation reason.
func (o *Orchestrator) HandleMaxIterations(wf *WorkflowInstance) (*WorkflowResult, error) {
	reason := fmt.Sprintf("maximum review iterations (%d) reached — unresolved findings require manual review", MaxIterations)
	return o.doEscalate(wf, reason)
}

// HandleAcceptanceRejection records an acceptance rejection decision
// on the workflow. Sets the accept stage to failed with the rejection
// rationale.
func (o *Orchestrator) HandleAcceptanceRejection(workflowID string, decision Decision) error {
	wf, err := o.store().Load(workflowID)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	// Find the accept stage and mark it as failed
	for i := range wf.Stages {
		if wf.Stages[i].StageName == StageAccept {
			now := o.now()
			wf.Stages[i].Status = StatusFailed
			wf.Stages[i].Error = fmt.Sprintf("rejected: %s", decision.Rationale)
			wf.Stages[i].CompletedAt = &now
			break
		}
	}

	now := o.now()
	wf.Status = StatusFailed
	wf.CompletedAt = &now

	return o.store().Save(wf)
}

// HandleContradiction escalates a workflow when two heroes produce
// contradictory guidance. Records both perspectives in the workflow
// state for human resolution.
func (o *Orchestrator) HandleContradiction(workflowID, conflict string) error {
	wf, err := o.store().Load(workflowID)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	now := o.now()
	wf.Status = StatusEscalated
	wf.CompletedAt = &now

	// Record the contradiction in the current stage
	if wf.CurrentStage >= 0 && wf.CurrentStage < len(wf.Stages) {
		wf.Stages[wf.CurrentStage].Status = StatusFailed
		wf.Stages[wf.CurrentStage].Error = fmt.Sprintf("contradiction: %s", conflict)
		wf.Stages[wf.CurrentStage].CompletedAt = &now
	}

	return o.store().Save(wf)
}

// effectiveMode returns the execution mode for a stage, defaulting to
// ModeHuman when the field is empty. This provides backward compatibility
// for legacy workflow JSON files that lack execution_mode fields (FR-010).
func effectiveMode(mode string) string {
	if mode == "" {
		return ModeHuman
	}
	return mode
}

// sanitizeBranch converts a branch name to a safe workflow ID component.
// Uses a restrictive allowlist: only letters, digits, and hyphens are kept.
// All other characters are replaced with hyphens.
func sanitizeBranch(branch string) string {
	var b strings.Builder
	for _, c := range branch {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			b.WriteRune(c)
		} else {
			b.WriteRune('-')
		}
	}
	return b.String()
}
