package schemas

// This file defines Go structs for all artifact payloads used in
// JSON Schema generation. Where existing Go structs exist in the
// codebase, we use type aliases to avoid duplication (DRY). For
// artifact types that don't have existing Go structs (quality-report,
// review-verdict), we define new structs here. All structs use
// json tags as the source of truth for schema generation via
// invopop/jsonschema.

import (
	"github.com/unbound-force/unbound-force/internal/artifacts"
	"github.com/unbound-force/unbound-force/internal/metrics"
	"github.com/unbound-force/unbound-force/internal/orchestration"
)

// --- Envelope Schema ---

// EnvelopeSchema represents the artifact envelope for JSON Schema
// generation. Uses typed Context and Payload fields instead of
// json.RawMessage to produce a fully-specified schema (FR-001).
type EnvelopeSchema struct {
	Hero          string        `json:"hero" jsonschema:"description=Producing hero identifier,example=gaze"`
	Version       string        `json:"version" jsonschema:"description=Hero version (semver),example=1.0.0"`
	Timestamp     string        `json:"timestamp" jsonschema:"description=ISO 8601 timestamp,example=2026-03-21T10:00:00Z"`
	ArtifactType  string        `json:"artifact_type" jsonschema:"description=Artifact type identifier,example=quality-report"`
	SchemaVersion string        `json:"schema_version" jsonschema:"description=Schema version (semver),example=1.0.0"`
	Context       ContextSchema `json:"context" jsonschema:"description=Workflow context metadata"`
	Payload       interface{}   `json:"payload" jsonschema:"description=Type-specific payload object"`
}

// ContextSchema represents the context field of the artifact
// envelope. Maps to artifacts.ArtifactContext but with schema
// metadata tags (FR-002, FR-014).
type ContextSchema struct {
	Branch        string `json:"branch,omitempty" jsonschema:"description=Git branch name"`
	Commit        string `json:"commit,omitempty" jsonschema:"description=Git commit SHA"`
	BacklogItemID string `json:"backlog_item_id,omitempty" jsonschema:"description=Originating backlog item ID"`
	CorrelationID string `json:"correlation_id,omitempty" jsonschema:"description=UUID linking related artifacts across workflow stages"`
	WorkflowID    string `json:"workflow_id,omitempty" jsonschema:"description=Workflow instance ID"`
}

// --- Type aliases for existing structs ---

// MetricsSnapshotPayload aliases the existing metrics.MetricsSnapshot
// struct. The existing struct already has json tags suitable for
// schema generation.
type MetricsSnapshotPayload = metrics.MetricsSnapshot

// CycleTimeStatsPayload aliases the existing metrics.CycleTimeStats.
type CycleTimeStatsPayload = metrics.CycleTimeStats

// BacklogHealthPayload aliases the existing metrics.BacklogHealth.
type BacklogHealthPayload = metrics.BacklogHealth

// HealthIndicatorPayload aliases the existing metrics.HealthIndicator.
type HealthIndicatorPayload = metrics.HealthIndicator

// WorkflowRecordPayload aliases the existing orchestration.WorkflowRecord.
type WorkflowRecordPayload = orchestration.WorkflowRecord

// WorkflowStagePayload aliases the existing orchestration.WorkflowStage.
type WorkflowStagePayload = orchestration.WorkflowStage

// DecisionPayload aliases the existing orchestration.Decision.
type DecisionPayload = orchestration.Decision

// AcceptanceDecisionPayload aliases the existing artifacts.AcceptanceDecision.
type AcceptanceDecisionPayload = artifacts.AcceptanceDecision

// ArtifactContextPayload aliases the existing artifacts.ArtifactContext.
type ArtifactContextPayload = artifacts.ArtifactContext

// --- New structs for types without existing Go definitions ---

// QualityReportPayload defines the payload for quality-report
// artifacts produced by Gaze. Mirrors the structure in the
// existing sample at schemas/samples/sample-quality-report-envelope.json.
type QualityReportPayload struct {
	Summary         QualityReportSummary    `json:"summary" jsonschema:"description=Overall quality scores"`
	Functions       []FunctionMetric        `json:"functions" jsonschema:"description=Per-function quality metrics"`
	Coverage        CoverageData            `json:"coverage" jsonschema:"description=Aggregate coverage data"`
	Recommendations []QualityRecommendation `json:"recommendations" jsonschema:"description=Improvement recommendations"`
}

// QualityReportSummary provides aggregate quality scores.
type QualityReportSummary struct {
	TotalFunctions int     `json:"total_functions" jsonschema:"description=Total number of analyzed functions"`
	AvgCoverage    float64 `json:"avg_coverage" jsonschema:"description=Average code coverage percentage"`
	AvgCrap        float64 `json:"avg_crap" jsonschema:"description=Average CRAP score"`
	CrapLoad       int     `json:"crap_load" jsonschema:"description=Number of functions with CRAP score above threshold"`
}

// FunctionMetric provides per-function quality data.
type FunctionMetric struct {
	Name             string  `json:"name" jsonschema:"description=Fully qualified function name"`
	CrapScore        float64 `json:"crap_score" jsonschema:"description=CRAP (Change Risk Anti-Patterns) score"`
	Complexity       int     `json:"complexity" jsonschema:"description=Cyclomatic complexity"`
	Coverage         float64 `json:"coverage" jsonschema:"description=Code coverage percentage"`
	ContractCoverage float64 `json:"contract_coverage" jsonschema:"description=Contract-level coverage percentage"`
	Classification   string  `json:"classification" jsonschema:"description=Function classification (contractual or incidental)"`
}

// CoverageData provides aggregate coverage statistics.
type CoverageData struct {
	TotalLines   int     `json:"total_lines" jsonschema:"description=Total lines of code"`
	CoveredLines int     `json:"covered_lines" jsonschema:"description=Lines covered by tests"`
	Percentage   float64 `json:"percentage" jsonschema:"description=Coverage percentage"`
}

// QualityRecommendation is an actionable improvement suggestion.
type QualityRecommendation struct {
	Priority    string `json:"priority" jsonschema:"description=Recommendation priority (high/medium/low)"`
	Description string `json:"description" jsonschema:"description=Human-readable recommendation"`
	Target      string `json:"target" jsonschema:"description=Target function or package"`
}

// ReviewVerdictPayload defines the payload for review-verdict
// artifacts produced by The Divisor.
type ReviewVerdictPayload struct {
	PersonaVerdicts    []PersonaVerdict `json:"persona_verdicts" jsonschema:"description=Individual persona review verdicts"`
	CouncilDecision    string           `json:"council_decision" jsonschema:"description=Overall council decision (APPROVED/CHANGES_REQUESTED/ESCALATED)"`
	IterationCount     int              `json:"iteration_count" jsonschema:"description=Number of review iterations"`
	UnresolvedFindings []ReviewFinding  `json:"unresolved_findings,omitempty" jsonschema:"description=Findings not yet addressed"`
	PRUrl              string           `json:"pr_url" jsonschema:"description=Pull request URL"`
	ConventionPackUsed string           `json:"convention_pack_used" jsonschema:"description=Convention pack ID used for review"`
}

// PersonaVerdict represents a single reviewer persona's assessment.
type PersonaVerdict struct {
	Persona  string          `json:"persona" jsonschema:"description=Reviewer persona name (guard/architect/adversary/sre/testing)"`
	Verdict  string          `json:"verdict" jsonschema:"description=Persona verdict (APPROVED/CHANGES_REQUESTED)"`
	Findings []ReviewFinding `json:"findings" jsonschema:"description=Specific findings from this persona"`
	Summary  string          `json:"summary" jsonschema:"description=Persona review summary"`
}

// ReviewFinding is a specific issue found during review.
type ReviewFinding struct {
	Severity    string `json:"severity" jsonschema:"description=Finding severity (CRITICAL/HIGH/MEDIUM/LOW)"`
	Category    string `json:"category" jsonschema:"description=Finding category"`
	Description string `json:"description" jsonschema:"description=Detailed finding description"`
	File        string `json:"file,omitempty" jsonschema:"description=Affected file path"`
	Line        int    `json:"line,omitempty" jsonschema:"description=Affected line number"`
}

// BacklogItemPayload defines the payload for backlog-item artifacts
// produced by Muti-Mind. Uses json tags (not yaml) because this
// represents the JSON artifact payload, not the Markdown frontmatter.
type BacklogItemPayload struct {
	ID                 string                `json:"id" jsonschema:"description=Backlog item identifier (e.g. BI-001)"`
	Title              string                `json:"title" jsonschema:"description=Item title"`
	Description        string                `json:"description,omitempty" jsonschema:"description=Detailed description"`
	Type               string                `json:"type" jsonschema:"description=Item type (feature/bug/chore/spike)"`
	Priority           string                `json:"priority" jsonschema:"description=Priority level (P1/P2/P3)"`
	Status             string                `json:"status" jsonschema:"description=Current status (backlog/ready/in-progress/done)"`
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria" jsonschema:"description=Acceptance criteria in Given/When/Then format"`
	Sprint             string                `json:"sprint,omitempty" jsonschema:"description=Assigned sprint"`
	EffortEstimate     string                `json:"effort_estimate,omitempty" jsonschema:"description=Effort estimate"`
	Dependencies       []string              `json:"dependencies,omitempty" jsonschema:"description=Dependent backlog item IDs"`
	RelatedSpecs       []string              `json:"related_specs,omitempty" jsonschema:"description=Related specification IDs"`
}

// AcceptanceCriterion represents a single Given/When/Then scenario.
type AcceptanceCriterion struct {
	Given string `json:"given" jsonschema:"description=Precondition"`
	When  string `json:"when" jsonschema:"description=Action"`
	Then  string `json:"then" jsonschema:"description=Expected outcome"`
}

// CoachingRecordPayload defines the payload for coaching-record
// artifacts produced by Mx F. Supports both retrospective and
// coaching interaction record types.
type CoachingRecordPayload struct {
	RecordType          string               `json:"record_type" jsonschema:"description=Record type (retrospective or coaching)"`
	Retrospective       *RetroPayload        `json:"retrospective,omitempty" jsonschema:"description=Retrospective session data"`
	CoachingInteraction *CoachingInteraction `json:"coaching_interaction,omitempty" jsonschema:"description=Coaching interaction data"`
}

// RetroPayload represents retrospective session data within a
// coaching record. Uses json tags for the artifact payload format
// (the existing coaching.RetroRecord uses yaml tags for Markdown
// frontmatter, so we define a separate struct for JSON output).
type RetroPayload struct {
	Date         string       `json:"date" jsonschema:"description=Retrospective date (YYYY-MM-DD)"`
	Participants []string     `json:"participants,omitempty" jsonschema:"description=Session participants"`
	Patterns     []string     `json:"patterns" jsonschema:"description=Identified patterns"`
	RootCauses   []string     `json:"root_causes,omitempty" jsonschema:"description=Root causes identified"`
	ActionItems  []ActionItem `json:"action_items" jsonschema:"description=Committed action items"`
}

// ActionItem represents a tracked improvement commitment in a
// coaching record payload.
type ActionItem struct {
	ID          string `json:"id" jsonschema:"description=Action item identifier"`
	Description string `json:"description" jsonschema:"description=Action item description"`
	Owner       string `json:"owner" jsonschema:"description=Responsible person or hero"`
	Deadline    string `json:"deadline" jsonschema:"description=Target completion date (YYYY-MM-DD)"`
	Status      string `json:"status" jsonschema:"description=Status (pending/in-progress/completed/stale)"`
}

// CoachingInteraction records a coaching session within a coaching
// record payload.
type CoachingInteraction struct {
	Topic     string   `json:"topic" jsonschema:"description=Coaching topic"`
	Questions []string `json:"questions" jsonschema:"description=Questions asked during session"`
	Insights  []string `json:"insights" jsonschema:"description=Insights surfaced"`
	Outcome   string   `json:"outcome" jsonschema:"description=Session outcome (action_item/escalation/resolved/deferred)"`
}
