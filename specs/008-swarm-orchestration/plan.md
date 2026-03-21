# Implementation Plan: Swarm Orchestration

**Branch**: `008-swarm-orchestration` | **Date**: 2026-03-20 | **Spec**: [spec.md](spec.md)

## Summary

Spec 008 implements the swarm orchestration layer: a Go package
(`internal/orchestration/`) that manages the hero lifecycle
workflow (define → implement → validate → review → accept →
measure), three OpenCode commands (`/workflow start|status|list`),
a Swarm skills package for the `opencode-swarm-plugin`, and
artifact protocol wiring to ensure all heroes produce/consume
artifacts at `.unbound-force/artifacts/`.

The orchestration engine is complementary to Swarm: our engine
handles domain workflow (hero stage sequencing, artifact handoff),
Swarm handles execution coordination (parallelism, checkpointing,
file reservations). Neither depends on the other.

## Technical Context

**Language/Version**: Go 1.24+ (orchestration engine), Markdown (commands, skills)
**Primary Dependencies**: `internal/artifacts` (envelope, FindArtifacts, WriteArtifact, ReadEnvelope — already exist from Spec 007), `internal/sync` (GHRunner), `github.com/charmbracelet/log`
**Storage**: JSON files at `.unbound-force/workflows/{workflow_id}.json` (workflow state), `.unbound-force/artifacts/{type}/{timestamp}-{hero}.json` (artifacts)
**Testing**: Standard library `testing`, `-race -count=1`, `t.TempDir()`
**Target Platform**: macOS, Linux
**Project Type**: Go package + OpenCode commands + Swarm skills
**Constraints**: No CLI binary for v1.0.0. OpenCode commands invoke the Go package. Swarm integration via skills packages only.

## Constitution Check

### I. Autonomous Collaboration — PASS

- Workflow stages communicate exclusively through artifacts
  at `.unbound-force/artifacts/`. No runtime coupling.
- Each stage produces a self-describing artifact envelope with
  hero, version, timestamp, artifact_type, and context.
- Workflow state is persisted as JSON — survives session death.
- Heroes can be invoked independently; orchestration is additive.

### II. Composability First — PASS

- FR-009: Each hero functions independently. The orchestration
  is additive — it enables the workflow but doesn't gate it.
- Removing a hero produces a degraded workflow with warnings,
  not errors (US5, FR-008).
- Swarm plugin is optional. `/workflow` commands work without
  Swarm. Manual hero invocation works without `/workflow`.

### III. Observable Quality — PASS

- `workflow-record` artifact captures the complete lifecycle:
  stages, artifacts, decisions, timing, outcome.
- All workflow state is JSON — machine-parseable.
- Workflow records are consumable by Mx F for metrics and by
  Muti-Mind for velocity tracking (FR-014).

### IV. Testability — PASS

- `Orchestrator` struct with injected dependencies (artifact
  store path, workflow store path, GHRunner, clock function).
- All state persisted to filesystem — testable with `t.TempDir()`.
- Hero availability detected by checking for agent files —
  mockable in tests.
- Stage transitions are pure functions on workflow state.
- Coverage target: 85% for `internal/orchestration/`.

**Gate Result**: ALL FOUR PRINCIPLES PASS.

## Project Structure

```text
# New orchestration package
internal/orchestration/
├── models.go          # WorkflowInstance, WorkflowStage, LearningFeedback, WorkflowRecord
├── engine.go          # Orchestrator struct, Start/Advance/Skip/Escalate/Complete methods
├── store.go           # Workflow state persistence (JSON at .unbound-force/workflows/)
├── heroes.go          # Hero availability detection, stage-to-hero mapping
├── record.go          # Workflow record generation (artifact production)
├── learning.go        # Learning feedback extraction and recommendation
├── engine_test.go     # Core engine tests (stage transitions, failure modes)
├── store_test.go      # Persistence round-trip tests
└── heroes_test.go     # Hero detection tests

# OpenCode commands (Markdown, not embedded in scaffold)
.opencode/command/
├── workflow-start.md   # /workflow start — begin a new workflow
├── workflow-status.md  # /workflow status — check active workflow
└── workflow-list.md    # /workflow list — list all workflows

# Swarm skills package
.opencode/skill/unbound-force-heroes/
└── SKILL.md            # Hero roles, routing patterns, workflow stages

# Artifact protocol updates
internal/artifacts/
└── artifacts.go        # Add ArtifactContext struct, update WriteArtifact signature
```

## Complexity Tracking

No constitution violations. The most complex area is the stage
transition state machine — 6 stages with failure/skip/escalate
transitions. Pure functions on state keep it testable.
<!-- scaffolded by unbound vdev -->
