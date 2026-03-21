# Quickstart: Swarm Orchestration

**Spec**: 008-swarm-orchestration
**Date**: 2026-03-20

## Prerequisites

- `unbound` CLI installed (`brew install unbound-force/tap/unbound`)
- At least one hero deployed (run `unbound init`)
- OpenCode running in the target project

## Start a Workflow

```
/workflow start BI-042
```

This creates a new hero lifecycle workflow scoped to the
current branch. The orchestrator:

1. Detects which heroes are available
2. Creates a 6-stage workflow (define → implement →
   validate → review → accept → measure)
3. Skips stages for unavailable heroes
4. Reports the first action to take

## Check Progress

```
/workflow status
```

Shows the current stage, completed stages, pending stages,
and artifacts produced so far.

## The Hero Lifecycle

```
1. /workflow start BI-042
   └─ Muti-Mind defines the work
      └─ /speckit.specify → /speckit.plan → /speckit.tasks

2. /workflow advance
   └─ Cobalt-Crush implements
      └─ /cobalt-crush (reads tasks.md, writes code)

3. /workflow advance
   └─ Gaze validates
      └─ /gaze (produces quality-report artifact)

4. /workflow advance
   └─ The Divisor reviews
      └─ /review-council (produces review-verdict artifact)
      └─ If REQUEST CHANGES → back to Cobalt-Crush (max 3x)

5. /workflow advance
   └─ Muti-Mind accepts
      └─ Produces acceptance-decision artifact

6. /workflow advance
   └─ Mx F measures
      └─ mxf collect → produces metrics-snapshot
      └─ Workflow record produced
```

## With Swarm Plugin (Optional)

If `opencode-swarm-plugin` is installed:

```
/swarm "Implement the health check endpoint from BI-042"
```

Swarm decomposes the task, spawns parallel workers, and
coordinates file reservations. Our workflow engine tracks
the hero lifecycle stages independently.

The Swarm skills package teaches the coordinator about
Unbound Force heroes:

```
skills_use({ name: "unbound-force-heroes" })
```

## Without Any Heroes

If no heroes are deployed, `/workflow start` reports:

```
No heroes available. Install heroes with:
  unbound init              # All heroes
  unbound init --divisor    # Review only
  brew install unbound-force/tap/gaze  # Quality analysis
```

## Graceful Degradation

Missing heroes are skipped automatically:

```
/workflow start BI-042

Available heroes:
  ✓ Cobalt-Crush (implement)
  ✓ The Divisor (review)
  ✗ Muti-Mind — not deployed (define, accept stages skipped)
  ✗ Gaze — not installed (validate stage skipped)
  ✗ Mx F — not installed (measure stage skipped)

Proceeding with 2/5 heroes. Skipped stages noted in workflow.
```

## List All Workflows

```
/workflow list
/workflow list --status completed
```
<!-- scaffolded by unbound vdev -->
