package artifacts

// ReviewVerdictDecision is a canonical review-verdict council decision.
type ReviewVerdictDecision string

const (
	// ReviewVerdictSchemaVersionV1 identifies historical review verdicts.
	ReviewVerdictSchemaVersionV1 = "1.0.0"
	// ReviewVerdictSchemaVersionV2 identifies current review verdicts.
	ReviewVerdictSchemaVersionV2 = "2.0.0"

	// ReviewVerdictApproved permits downstream automated progression.
	ReviewVerdictApproved ReviewVerdictDecision = "APPROVED"
	// ReviewVerdictChangesRequested blocks progression on review findings.
	ReviewVerdictChangesRequested ReviewVerdictDecision = "CHANGES_REQUESTED"
	// ReviewVerdictEscalated blocks progression pending human resolution.
	ReviewVerdictEscalated ReviewVerdictDecision = "ESCALATED"
	// ReviewVerdictInconclusive blocks progression when no result was calculated.
	ReviewVerdictInconclusive ReviewVerdictDecision = "INCONCLUSIVE"
	// ReviewVerdictUnavailable blocks progression when assessment was unavailable.
	ReviewVerdictUnavailable ReviewVerdictDecision = "UNAVAILABLE"
)
