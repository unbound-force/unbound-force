# Quickstart: Swarm Delegation Workflow

**Date**: 2026-03-22
**Branch**: `012-swarm-delegation`

## Overview

After this change, the hero lifecycle workflow supports
autonomous swarm delegation. The human specifies and
clarifies a feature, then the swarm runs implementation
through review autonomously. The workflow pauses for the human to
accept, then the swarm runs the final reflect stage.

## Typical Session

### Step 1: Human defines the feature

```
> /speckit.specify
> /speckit.clarify
> /workflow start BI-042
```

Output:

```
Workflow started: wf-feat-login-20260322T143000

Available heroes:
  ✓ Muti-Mind (define) [human]
  ✓ Cobalt-Crush (implement) [swarm]
  ✓ Gaze (validate) [swarm]
  ✓ The Divisor (review) [swarm]
  ✓ Muti-Mind (accept) [human]
  ✓ Mx F (reflect) [swarm]

Stage 1/6: define (Muti-Mind) [human]
```

### Step 2: Human completes define and hands off

```
> /workflow advance
```

Output:

```
Stage advanced: define → implement

Stage 2/6: implement (Cobalt-Crush) [swarm]
  Next: Swarm will implement from tasks.md.
```

The swarm now takes over. The Swarm coordinator
automatically calls `/workflow advance` for each
swarm-mode stage (implement, validate, review) without
human intervention. See the `unbound-force-heroes`
SKILL.md for how the coordinator drives swarm stages.

### Step 3: Swarm completes, workflow pauses

After the swarm finishes all its stages, the workflow
pauses automatically:

```
Workflow paused: wf-feat-login-20260322T143000

Swarm completed: implement → validate → review
Awaiting human action at: accept (Muti-Mind) [human]

Run /workflow advance to resume.
```

### Step 4: Human accepts

```
> /workflow advance
```

Output:

```
Workflow resumed: wf-feat-login-20260322T143000

Stage 5/6: accept (Muti-Mind) [human]
  Next: Review the increment and make acceptance decision.
```

The human reviews the Gaze quality report, checks the
Divisor's review verdict, and decides whether to accept.

```
> /workflow advance
```

### Step 5: Swarm runs reflect automatically

The swarm runs the final reflect stage (Mx F) which
collects metrics, runs learning analysis, and produces
a retrospective summary.

```
Workflow completed: wf-feat-login-20260322T143000

All stages:
  ✓ define      (muti-mind)     15m   [human]
  ✓ implement   (cobalt-crush)  2h    [swarm]
  ✓ validate    (gaze)          5m    [swarm]
  ✓ review      (divisor)       20m   [swarm]
  ✓ accept      (muti-mind)     10m   [human]
  ✓ reflect     (mx-f)          5m    [swarm]

Total time: 3h15m
Outcome: shipped
```

## Human Interactions

The complete workflow requires exactly 2 human
interactions:

1. **After define**: `/workflow advance` to hand off to
   the swarm
2. **At accept**: `/workflow advance` to resume, review,
   and accept

Everything else runs autonomously.

## Checking Status

At any point, run `/workflow status` to see where the
workflow stands:

```
> /workflow status
```

If the workflow is paused, the status clearly shows
which stage needs human attention:

```
Status: awaiting_human

  ⏸ accept (muti-mind) pending [human]  ← your turn
```

## Legacy Workflows

Existing workflows created before this change continue
to work as before. They advance one stage at a time with
no automatic pausing.
