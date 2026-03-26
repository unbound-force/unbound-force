# Quickstart: Autonomous Define with Dewey

**Date**: 2026-03-26
**Branch**: `016-autonomous-define`

## Overview

After this change, the hero lifecycle workflow supports
a single-checkpoint mode. The human expresses intent
with one sentence (seed), the swarm handles everything
(define through review), and the human reviews the
complete increment for acceptance.

## The Seed-to-Accept Workflow

### Step 1: Seed

```
> /workflow seed "add CSV export to the dashboard"
```

That's it. One sentence. Muti-Mind takes over.

### Step 2: Wait

The swarm autonomously:
1. **Muti-Mind** drafts the spec using Dewey context
2. **Cobalt-Crush** plans and implements
3. **Gaze** validates quality
4. **The Divisor** reviews the code

### Step 3: Accept

```
Workflow paused: wf-feat-csv-export-20260326T143000

Swarm completed: define → implement → validate → review
Awaiting human action at: accept (Muti-Mind) [human]

Run /workflow advance to accept the increment.
```

The human reviews:
- The specification (was the intent captured correctly?)
- The Gaze quality report
- The Divisor review verdict

```
> /workflow advance
```

### Step 4: Done

The swarm runs the reflect stage (Mx F metrics +
learning), and the workflow completes.

## With Spec Review Checkpoint

For high-stakes features, add `--spec-review`:

```
> /workflow seed "migrate all user credentials to OAuth2" --spec-review
```

The workflow pauses twice:
1. After define -- human reviews the spec
2. Before accept -- human reviews the implementation

```
Specification drafted: wf-feat-oauth2-20260326T150000

Review at: specs/NNN-oauth2-migration/spec.md
Run /workflow advance to approve the spec.

> /workflow advance

Spec approved. Continuing to implement...
```

## Configuration

### Using `/workflow start` with Mode Override

```
> /workflow start BI-042 --define-mode=swarm
```

This starts a workflow with the define stage in swarm
mode. All other stages keep their defaults.

### Default Behavior (Unchanged)

```
> /workflow start BI-042
```

This starts a workflow with the define stage in human
mode -- exactly the same behavior as before this change.

## Comparison

| Workflow | Human Checkpoints | Commands |
|----------|:-----------------:|---------|
| **Default** (define=human) | 2 | specify + clarify + advance, then accept |
| **Autonomous** (define=swarm) | 1 | seed, then accept |
| **Autonomous + spec review** | 2 | seed, review spec, then accept |
