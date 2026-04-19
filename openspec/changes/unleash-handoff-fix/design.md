## Context

After `/unleash` was implemented (Spec 018) and extended
to support OpenSpec branches (Spec 031), it became the
primary autonomous pipeline. However, the handoff
prompts in spec-phase commands and skills were not
updated consistently. Some omit `/unleash` entirely,
others list it last (alphabetical ordering rather than
priority ordering).

The canonical implementation command priority is:

1. `/unleash` — autonomous pipeline with parallel swarm
   workers (recommended for multi-task changes)
2. `/cobalt-crush` — developer agent, auto-detects
   workflow context (good for single-agent work)
3. `/opsx-apply` — sequential single-agent (OpenSpec)
4. `/speckit.implement` — sequential single-agent
   (Speckit)

## Goals / Non-Goals

### Goals

- Every spec-phase STOP block lists `/unleash` first
- Every output prompt recommends `/unleash` as primary
- Every guardrails "NEVER run" list includes `/unleash`
- The `/cobalt-crush` fallback prompt includes `/unleash`
- The `uf-init.md` template propagates correct text
- Fix the `/opsx:apply` typo in `cobalt-crush.md`

### Non-Goals

- Changing Go source code or test files
- Modifying agent persona files (divisor-*, cobalt-crush-dev, etc.)
- Modifying convention packs
- Fixing stale Hivemind references (separate issue)
- Fixing the self-referential guardrails in `speckit.implement.md`
  (that's a different bug — the guardrails say "NEVER modify
  source code" on the command that IS supposed to modify source
  code; needs its own investigation)

## Decisions

**D1: Canonical ordering** — All command lists use the
order `/unleash`, `/cobalt-crush`, `/opsx-apply`. This
puts the most autonomous option first and the most
manual last.

**D2: Scaffold asset sync required for 3 files** — Most
of the affected command files are NOT embedded scaffold
assets (the spec-phase `speckit.*.md` commands were
externalized in Spec 027). However, 3 of the 13 target
files ARE embedded assets that require sync:
- `.opencode/command/cobalt-crush.md` →
  `internal/scaffold/assets/opencode/command/cobalt-crush.md`
- `.opencode/command/uf-init.md` →
  `internal/scaffold/assets/opencode/command/uf-init.md`
- `.opencode/skill/speckit-workflow/SKILL.md` →
  `internal/scaffold/assets/opencode/skill/speckit-workflow/SKILL.md`

These must be synced after editing the live files, or
the `TestScaffoldOutput_*` drift detection tests will
fail CI.

**D3: `uf-init.md` template must match** — The
`uf-init.md` command contains a guardrails template
block that gets injected into `opsx-propose.md` during
`/uf-init`. This template must be updated to match the
corrected text, otherwise fresh `/uf-init` runs would
re-inject the old text.

**D4: `speckit-workflow/SKILL.md` gets an entry point
section** — This skill describes how Swarm coordinators
work with Speckit tasks but never mentions `/unleash` as
the trigger command. Adding a brief Entry Point section
connects the skill to the command.

## Risks / Trade-offs

- **Risk**: Agents may over-recommend `/unleash` for
  trivial 1-2 task changes where `/opsx-apply` would be
  faster. **Mitigation**: The handoff text says
  "recommended for multi-task changes" — agents should
  use judgment.

- **Trade-off**: Listing `/unleash` first in STOP blocks
  could confuse users who don't have Replicator installed
  (Swarm worktrees unavailable). **Mitigation**:
  `/unleash` gracefully degrades to sequential execution
  when Swarm is unavailable, so it still works — just
  without parallelism.
