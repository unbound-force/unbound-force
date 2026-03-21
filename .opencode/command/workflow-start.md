---
name: workflow-start
description: Begin a new hero lifecycle workflow
---

# /workflow start

Begin a new hero lifecycle workflow for the current feature branch.

## Usage

```
/workflow start [backlog-item-id]
```

## Behavior

When this command is invoked:

1. **Detect the current git branch** by running `git branch --show-current`.

2. **Check for existing active workflows** by reading JSON files from `.unbound-force/workflows/`. If an active workflow already exists for this branch, report it and ask if the user wants to create a new one.

3. **Detect available heroes** by checking:
   - `.opencode/agents/muti-mind-po.md` → Muti-Mind (define, accept)
   - `.opencode/agents/cobalt-crush-dev.md` → Cobalt-Crush (implement)
   - Any `.opencode/agents/divisor-*.md` → The Divisor (review)
   - `.opencode/agents/mx-f-coach.md` → Mx F (measure)
   - `which gaze` → Gaze (validate)

4. **Create a new workflow JSON file** at `.unbound-force/workflows/{workflow_id}.json` with:
   - `workflow_id`: `wf-{branch}-{timestamp}` (e.g., `wf-feat-health-check-20260320T143000`)
   - `feature_branch`: current git branch
   - `backlog_item_id`: from argument (or empty)
   - `stages`: 6 stages (define, implement, validate, review, accept, measure)
   - `status`: "active"
   - Mark unavailable hero stages as "skipped" with reason

5. **Activate the first non-skipped stage**.

6. **Report the result** showing:
   - Workflow ID
   - Available heroes with checkmarks (✓) and unavailable with crosses (✗)
   - Current stage and next action

## Output Format

```
Workflow started: wf-feat-health-check-20260320T143000

Available heroes:
  ✓ Muti-Mind (define)
  ✓ Cobalt-Crush (implement)
  ✗ Gaze (validate) — not installed
  ✓ The Divisor (review)
  ✓ Muti-Mind (accept)
  ✓ Mx F (measure)

Stage 1/6: define (Muti-Mind)
  Next: Run /speckit.specify to create the feature spec.
```

## Directory Structure

```
.unbound-force/
└── workflows/
    └── wf-feat-health-check-20260320T143000.json
```
