---
name: workflow-seed
description: Seed a new feature workflow with one sentence
---

# /workflow seed

Start a new hero lifecycle workflow from a one-sentence
feature description. Combines backlog item creation with
workflow start in one operation, with the define stage
running in swarm mode.

## Usage

```
/workflow seed <description> [--spec-review]
```

### Flags

- `--spec-review`: Enable the spec review checkpoint
  between define and implement. When enabled, the workflow
  pauses after Muti-Mind drafts the spec, allowing the
  human to review before implementation begins.

## Behavior

When this command is invoked:

1. **Validate the description**: If `<description>` is
   empty, prompt the operator for a feature description
   before proceeding.

2. **Create a backlog item** from the seed description:
   ```bash
   mutimind add --title "<description>" --type story
   ```

3. **Start a workflow** with the define stage in swarm
   mode:
   - Call the orchestration engine's `Start()` with
     `overrides={"define": "swarm"}` and the
     `specReview` flag from the `--spec-review` option.
   - The workflow ID is generated automatically.

4. **Report the result** showing the workflow ID, stage
   layout, and next action.

## Output Format

### Successful Seed

```
Seeded: wf-feat-csv-export-20260326T143000

Muti-Mind is drafting the specification...

Workflow stages:
  ◉ define      (muti-mind)     active   [swarm]
  ○ implement   (cobalt-crush)  pending  [swarm]
  ○ validate    (gaze)          pending  [swarm]
  ○ review      (divisor)       pending  [swarm]
  ○ accept      (muti-mind)     pending  [human]
  ○ reflect     (mx-f)          pending  [swarm]

The swarm will notify you when the increment is
ready for acceptance.
```

### Seed with Spec Review

```
Seeded: wf-feat-oauth2-20260326T150000

Muti-Mind is drafting the specification...
Spec review checkpoint enabled.

Workflow stages:
  ◉ define      (muti-mind)     active   [swarm]
  ○ implement   (cobalt-crush)  pending  [swarm]
  ○ validate    (gaze)          pending  [swarm]
  ○ review      (divisor)       pending  [swarm]
  ○ accept      (muti-mind)     pending  [human]
  ○ reflect     (mx-f)          pending  [swarm]

The workflow will pause after the spec is drafted
for your review.
```

### Empty Description

```
> /workflow seed

Please provide a feature description:
> add CSV export to the dashboard

Seeded: wf-feat-csv-export-20260326T143000
...
```

## Comparison with /workflow start

| Command | Define Mode | Human Steps |
|---------|:-----------:|:-----------:|
| `/workflow start BI-042` | human (default) | specify + clarify + advance, then accept |
| `/workflow start BI-042 --define-mode=swarm` | swarm | seed, then accept |
| `/workflow seed "description"` | swarm (always) | seed, then accept |
| `/workflow seed "description" --spec-review` | swarm (always) | seed, review spec, then accept |

The seed command is a convenience wrapper around
`/workflow start` with `--define-mode=swarm`. It also
creates the backlog item automatically.
