package orchestration

import "time"

// GenerateWorkflowRecord extracts a complete WorkflowRecord from
// a WorkflowInstance. It computes total elapsed time, collects all
// artifact paths, and determines the outcome based on stage statuses.
//
// Outcome logic:
//   - shipped: all non-skipped stages completed
//   - rejected: acceptance stage has a rejection decision (failed status)
//   - abandoned: workflow failed or was escalated
func GenerateWorkflowRecord(wf *WorkflowInstance, now time.Time) *WorkflowRecord {
	record := &WorkflowRecord{
		WorkflowID:    wf.WorkflowID,
		BacklogItemID: wf.BacklogItemID,
		Stages:        wf.Stages,
	}

	// Collect all artifact paths from stages
	var allArtifacts []string
	for _, stage := range wf.Stages {
		allArtifacts = append(allArtifacts, stage.ArtifactsProduced...)
	}
	record.Artifacts = allArtifacts

	// Compute total elapsed time
	elapsed := now.Sub(wf.StartedAt)
	record.TotalElapsedTime = elapsed.String()

	// Determine outcome
	record.Outcome = determineOutcome(wf)

	return record
}

// determineOutcome evaluates the workflow state to determine the
// final outcome. Uses a priority-based approach: escalated/failed
// takes precedence, then rejection, then shipped.
func determineOutcome(wf *WorkflowInstance) string {
	if wf.Status == StatusEscalated || wf.Status == StatusFailed {
		return OutcomeAbandoned
	}

	// Check if acceptance stage was rejected (failed)
	for _, stage := range wf.Stages {
		if stage.StageName == StageAccept && stage.Status == StatusFailed {
			return OutcomeRejected
		}
	}

	// All non-skipped stages completed
	return OutcomeShipped
}
