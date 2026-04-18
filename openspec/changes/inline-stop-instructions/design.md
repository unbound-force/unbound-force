## Context

Spec-phase commands (`/opsx-propose`, `/speckit.clarify`,
etc.) have `## Guardrails` sections at the bottom that
say "NEVER implement." But the agent reads the workflow
steps sequentially, completes the work, then has
momentum to continue. The guardrails at the end are
too late — the agent has already moved on.

## Goals / Non-Goals

### Goals

- Add an inline STOP instruction at the exact
  completion point of each spec-phase command
- Add "why" reasoning to existing Guardrails sections
- Add implementation prevention guardrails to the
  openspec-propose skill (currently only has artifact
  quality guardrails)

### Non-Goals

- No changes to `/speckit.implement` (implementation
  IS its job)
- No changes to `/speckit.constitution` (governance,
  not implementation risk)
- No changes to `/speckit.taskstoissues` (creates GH
  issues, low risk)
- No Go code changes
- No scaffold asset syncs (target files are externalized)

## Decisions

### D1: 3-part fix per file

Each file gets:
1. **Inline STOP** — after the command's work is done
2. **"Why" in Guardrails** — reasoning added to
   existing guardrails
3. **Output reinforcement** — "CRITICAL: you are DONE"
   in the output/report section (where applicable)

### D2: Inline STOP text (standardized)

All files use the same STOP block:

```markdown
**STOP HERE. Do NOT proceed to implementation.**

Your job is done. Report the results and prompt the
user. The user will invoke a separate command
(/opsx-apply, /cobalt-crush, or /unleash) when they
are ready to implement.
```

### D3: "Why" text (standardized)

All guardrails add this reasoning:

```markdown
The user needs to review the plan before
implementation begins. Implementing without review
defeats the purpose of the spec-first workflow.
```

### D4: Skill file gets full guardrails

The openspec-propose skill currently has artifact
quality guardrails but NOT implementation prevention
guardrails. Add the same "NEVER implement" block that
the command file has.
