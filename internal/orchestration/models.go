// Package orchestration implements the swarm orchestration engine
// for the Unbound Force hero lifecycle workflow. It manages the
// 6-stage feature lifecycle (define, implement, validate, review,
// accept, reflect), hero availability detection, workflow state
// persistence, and learning feedback extraction.
package orchestration

import "time"

// Stage name constants define the 6 stages of the hero lifecycle.
const (
	StageDefine    = "define"
	StageImplement = "implement"
	StageValidate  = "validate"
	StageReview    = "review"
	StageAccept    = "accept"
	StageReflect   = "reflect"
)

// StageOrder returns the canonical sequence of workflow stages.
// Returned as a fresh slice to prevent mutation of the canonical order.
func StageOrder() []string {
	return []string{
		StageDefine,
		StageImplement,
		StageValidate,
		StageReview,
		StageAccept,
		StageReflect,
	}
}

// Status constants for workflow and stage state transitions.
const (
	StatusPending       = "pending"
	StatusActive        = "active"
	StatusCompleted     = "completed"
	StatusSkipped       = "skipped"
	StatusFailed        = "failed"
	StatusEscalated     = "escalated"
	StatusAwaitingHuman = "awaiting_human"
)

// Execution mode constants indicate whether a stage is driven by
// a human operator or by the swarm autonomously.
const (
	ModeHuman = "human"
	ModeSwarm = "swarm"
)

// Outcome constants for workflow records.
const (
	OutcomeShipped   = "shipped"
	OutcomeRejected  = "rejected"
	OutcomeAbandoned = "abandoned"
)

// WorkflowInstance represents a single execution of the hero
// lifecycle for a feature. Persisted as JSON at
// .unbound-force/workflows/{workflow_id}.json.
type WorkflowInstance struct {
	WorkflowID      string          `json:"workflow_id"`
	FeatureBranch   string          `json:"feature_branch"`
	BacklogItemID   string          `json:"backlog_item_id"`
	Stages          []WorkflowStage `json:"stages"`
	CurrentStage    int             `json:"current_stage"`
	StartedAt       time.Time       `json:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	Status          string          `json:"status"`
	AvailableHeroes []string        `json:"available_heroes"`
	IterationCount  int             `json:"iteration_count"`
}

// WorkflowStage represents one step in the hero lifecycle.
type WorkflowStage struct {
	StageName         string     `json:"stage_name"`
	Hero              string     `json:"hero"`
	Status            string     `json:"status"`
	ExecutionMode     string     `json:"execution_mode,omitempty"`
	ArtifactsProduced []string   `json:"artifacts_produced,omitempty"`
	ArtifactsConsumed []string   `json:"artifacts_consumed,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	SkipReason        string     `json:"skip_reason,omitempty"`
	Error             string     `json:"error,omitempty"`
}

// Decision represents a decision point in the workflow
// (review verdict or acceptance decision).
type Decision struct {
	Type      string    `json:"type"`
	Hero      string    `json:"hero"`
	Result    string    `json:"result"`
	Rationale string    `json:"rationale"`
	Iteration int       `json:"iteration"`
	Timestamp time.Time `json:"timestamp"`
}

// LearningFeedback represents a cross-hero recommendation
// produced by analyzing completed workflow records.
type LearningFeedback struct {
	ID              string            `json:"id"`
	SourceHero      string            `json:"source_hero"`
	TargetHero      string            `json:"target_hero"`
	PatternObserved string            `json:"pattern_observed"`
	Recommendation  string            `json:"recommendation"`
	SupportingData  map[string]string `json:"supporting_data,omitempty"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	WorkflowIDs     []string          `json:"workflow_ids"`
}

// WorkflowRecord captures the complete lifecycle history of a
// feature workflow. Produced as an artifact on workflow completion.
type WorkflowRecord struct {
	WorkflowID       string          `json:"workflow_id"`
	BacklogItemID    string          `json:"backlog_item_id"`
	Stages           []WorkflowStage `json:"stages"`
	Artifacts        []string        `json:"artifacts"`
	Decisions        []Decision      `json:"decisions"`
	TotalElapsedTime string          `json:"total_elapsed_time"`
	Outcome          string          `json:"outcome"`
	LearningFeedback []string        `json:"learning_feedback,omitempty"`
}

// HeroStatus represents the availability state of a hero,
// detected at workflow start.
type HeroStatus struct {
	Name            string `json:"name"`
	Role            string `json:"role"`
	Available       bool   `json:"available"`
	AgentFile       string `json:"agent_file"`
	DetectionMethod string `json:"detection_method"`
}
