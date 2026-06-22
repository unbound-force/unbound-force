## Context

The `/uf-init` slash command (`.opencode/commands/uf-init.md`)
applies project-specific customizations to OpenSpec skill files
and speckit command files. Cross-repo comparison between
`unbound-force` and `gaze` revealed that several customizations
present in the meta repo are never applied to downstream repos
because they are not defined in any `/uf-init` step.

The command currently has 10 steps (0-9, where Step 9 is
the report template). This change adds 3 new steps
(10-12) and modifies 2 existing steps (2 and 6),
bringing the total to 13 steps (0-12).

## Goals / Non-Goals

### Goals
- Every customization present in `unbound-force/.opencode/`
  files is traceable to a specific `/uf-init` step
- Idempotency checks are specific enough to distinguish
  between partial and complete application of each
  customization category
- Legacy directory artifacts are cleaned up
- Scaffold comment accumulation is prevented

### Non-Goals
- Modifying the Go `uf init` binary (this change is
  command-only)
- Changing the content of the customizations themselves
  (we are codifying what already exists in unbound-force)
- Addressing issues #161/#162/#201/#213 (those relate to
  the Go binary, not the slash command)
- Verifying content parity between `uf/packs/` and
  `unbound/packs/` before cleanup (the migration is
  already complete in all active repos)

## Decisions

### D1: Add steps rather than restructure existing ones

**Decision**: Add new steps (10, 11, 12) rather than
restructuring the existing step numbering.

**Rationale**: The existing step numbers (0-9) are
referenced in AGENTS.md and potentially in user memory.
Renumbering would create unnecessary confusion. Appending
new steps at the end maintains backward compatibility with
partial-run expectations.

### D2: Split Guardrails into spec-phase and execution-phase

**Decision**: Modify Step 6 to inject a variant Guardrails
block depending on whether the command is a spec-phase
command (specify, plan, tasks, analyze, checklist, clarify)
or an execution/utility command (implement, constitution,
taskstoissues).

**Rationale**: The review-rationale sentence ("The user
needs to review the plan before implementation begins...")
is only meaningful for spec-phase commands. Adding it to
implement or constitution commands would be contradictory.

### D3: Tighten branch enforcement idempotency via specific markers

**Decision**: Instead of checking semantically for "any
branch management content," check for specific marker
phrases that distinguish variants:
- Basic branch check: `opsx/<name>` or `opsx/<change-name>`
- Dirty tree check: `git status --short` in a
  pre-branch-creation context
- Commit-before-archive: `git add` + `git commit` before
  the archive move step
- Branch-switch confirmation: `uncommitted changes` in
  explore guardrails

**Rationale**: The current semantic check ("does the file
already describe creating, validating, or cleaning up an
opsx branch?") is too coarse. It detects the basic branch
check and concludes the entire branch enforcement category
is present, skipping the enhanced variants.

### D4: STOP HERE blocks as a separate step

**Decision**: Create a dedicated step for STOP HERE blocks
rather than folding them into the existing Guardrails step.

**Rationale**: STOP HERE blocks serve a different purpose
than Guardrails. Guardrails constrain what the command may
do; STOP HERE blocks control workflow flow (preventing
premature advancement to implementation). They also have
different insertion points (after the main workflow, before
Guardrails vs. appended at the end).

### D5: Scaffold comment deduplication approach

**Decision**: Use a simple "keep the last one, remove
earlier duplicates" strategy for scaffold comments. Match
the pattern `<!-- scaffolded by uf ... -->`.

**Rationale**: The most recent scaffold comment carries the
version information that matters. Earlier ones are
archaeological artifacts of previous `uf init` runs.

## Risks / Trade-offs

### R1: Step count growth

The command grows from 9 to 12 substantive steps. This
increases execution time and LLM context consumption.

**Mitigation**: Each new step follows the established
pattern (read, check, skip-or-insert, report) and adds
minimal instruction text. The idempotency checks mean
subsequent runs of already-customized files skip quickly.

### R2: Marker-based idempotency is fragile

Checking for specific strings (D3) could break if the
customization text is reworded in a future update.

**Mitigation**: The markers chosen are structural
identifiers unlikely to change accidentally
(`git status --short`, `git add`, `uncommitted changes`).
If the customization text is intentionally reworded, the
marker strings in `/uf-init` should be updated in the same
change.

### R3: Step count growth and maintenance burden

The command grows from 10 to 13 steps. Steps 11 and 12
are time-limited cleanup steps that should be retired
once all downstream repos are clean:

- **Step 11** (Scaffold Comment Dedup): Retire after all
  active repos report `⊘` for all files. Estimated: 1-2
  release cycles after this change ships.
- **Step 12** (Legacy Directory Cleanup): Retire after
  all active repos report `⊘ unbound/packs/: not
  present`. Estimated: 1-2 release cycles.

The root cause of scaffold comment accumulation was fixed
in PR #127 (idempotent marker insertion in the Go
binary). Step 11 remediates existing accumulation.

### R4: Legacy cleanup could remove needed files

Removing `unbound/packs/` in a downstream repo that
somehow still depends on it would break that repo.

**Mitigation**: The cleanup step verifies `uf/packs/`
exists and contains at least the core files
(`default.md`, `severity.md`) before removing
`unbound/packs/`. If `uf/packs/` is missing or
incomplete, the step reports an error and skips cleanup.
