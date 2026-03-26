---
name: workflow-start
description: Begin a new hero lifecycle workflow
---

# /workflow start

Begin a new hero lifecycle workflow for the current feature branch.

## Usage

```
/workflow start [backlog-item-id] [--define-mode=human|swarm] [--spec-review]
```

### Flags

- `--define-mode`: Set the define stage execution mode.
  Accepts `human` (default) or `swarm`. When set to `swarm`,
  Muti-Mind autonomously drafts the specification using
  Dewey context without human interaction.
- `--spec-review`: Enable the spec review checkpoint between
  define and implement. When enabled AND define is in swarm
  mode, the workflow pauses after Muti-Mind completes the
  spec, allowing the human to review before implementation
  begins. Disabled by default.

## Behavior

When this command is invoked:

1. **Detect the current git branch** by running `git branch --show-current`.

2. **Check for existing active or awaiting_human workflows** by reading JSON files from `.unbound-force/workflows/`. If an in-progress workflow already exists for this branch, report it and ask if the user wants to create a new one.

3. **Detect available heroes** by checking:
   - `.opencode/agents/muti-mind-po.md` → Muti-Mind (define, accept)
   - `.opencode/agents/cobalt-crush-dev.md` → Cobalt-Crush (implement)
   - Any `.opencode/agents/divisor-*.md` → The Divisor (review)
   - `.opencode/agents/mx-f-coach.md` → Mx F (reflect)
   - `which gaze` → Gaze (validate)

4. **Create a new workflow JSON file** at `.unbound-force/workflows/{workflow_id}.json` with:
   - `workflow_id`: `wf-{branch}-{timestamp}` (e.g., `wf-feat-health-check-20260320T143000`)
   - `feature_branch`: current git branch
   - `backlog_item_id`: from argument (or empty)
   - `stages`: 6 stages (define, implement, validate, review, accept, reflect)
   - `status`: "active"
   - Each stage includes `execution_mode`: `"human"` or `"swarm"`
   - Mark unavailable hero stages as "skipped" with reason

5. **Activate the first non-skipped stage**.

6. **Report the result** showing:
   - Workflow ID
   - Available heroes with checkmarks (✓) and unavailable with crosses (✗), including `[human]`/`[swarm]` mode indicators
   - Current stage and next action

## Output Format

```
Workflow started: wf-feat-health-check-20260320T143000

Available heroes:
  ✓ Muti-Mind (define) [human]
  ✓ Cobalt-Crush (implement) [swarm]
  ✗ Gaze (validate) [swarm] — not installed
  ✓ The Divisor (review) [swarm]
  ✓ Muti-Mind (accept) [human]
  ✓ Mx F (reflect) [swarm]

Stage 1/6: define (Muti-Mind) [human]
  Next: Run /speckit.specify to create the feature spec.
```

## Autonomous Define Example

```
> /workflow start BI-042 --define-mode=swarm

Workflow started: wf-feat-csv-export-20260326T143000

Available heroes:
  ✓ Muti-Mind (define) [swarm]
  ✓ Cobalt-Crush (implement) [swarm]
  ✓ Gaze (validate) [swarm]
  ✓ The Divisor (review) [swarm]
  ✓ Muti-Mind (accept) [human]
  ✓ Mx F (reflect) [swarm]

Stage 1/6: define (Muti-Mind) [swarm]
  Muti-Mind is drafting the specification autonomously...
```

## Autonomous Define with Spec Review Example

```
> /workflow start BI-042 --define-mode=swarm --spec-review

Workflow started: wf-feat-oauth2-20260326T150000

Stage 1/6: define (Muti-Mind) [swarm]
  Muti-Mind is drafting the specification autonomously...
  Spec review checkpoint enabled — workflow will pause after define.
```

## Directory Structure

```
.unbound-force/
└── workflows/
    └── wf-feat-health-check-20260320T143000.json
```
