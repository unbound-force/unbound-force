## Context

`uf.init.md` Step 6 injects a `## Guardrails` section into all
9 speckit command files. The current implementation uses two
template variants that differ only by the presence of a
review-rationale sentence — the restrictive content ("NEVER
modify source code", "ONLY files this command may write are
FEATURE_SPEC/FEATURE_DIR") is identical in both.

This is factually wrong for 3 execution/utility commands whose
operational models differ from spec-artifact authoring. A
workaround Note (lines 627–632) acknowledges the `implement`
contradiction but relies on "instructions override guardrails."

Both `.opencode/commands/uf.init.md` and
`internal/scaffold/assets/opencode/commands/uf.init.md` are
byte-identical and must be updated in lockstep. The drift
detection test (`TestEmbeddedAssets_MatchSource`) enforces this.

## Goals / Non-Goals

### Goals
- Replace the shared execution/utility guardrail template with
  3 command-specific guardrail blocks that accurately describe
  each command's operational model
- Add idempotency logic that detects and replaces incorrect
  guardrails (not just injects when absent)
- Remove the workaround Note (lines 627–632) that documents
  the implement override conflict
- Add guardrail content regression tests to prevent recurrence
- Keep both file copies (deployed + scaffold asset) in sync

### Non-Goals
- Preset migration (#257) — that is a strategic architectural
  change to the guardrail delivery mechanism, not this fix
- Upstream speckit changes (#258) — separate proposal
- Fixing `speckit.testreview.md` guardrail — its read-only
  constraints are stricter than the guardrail, making the
  error harmless (noted for future cleanup)
- Changing the 6 spec-phase command guardrails — those are
  correct as-is

## Decisions

### D1: Three command-specific variants, not one parameterized template

**Decision**: Define 3 separate markdown blocks — one for each
execution/utility command — rather than a parameterized template
with placeholder substitution.

**Rationale**: Each command has a fundamentally different
operational model (source code writes vs. config file writes
vs. API calls). A parameterized template would need so many
conditionals that it would be harder to read than distinct
blocks. The existing Step 6 structure already uses distinct
markdown code blocks; adding 3 more follows the pattern.

### D2: Replace-if-incorrect idempotency

**Decision**: When `## Guardrails` already exists but contains
incorrect content (e.g., "NEVER modify source code" in
`speckit.implement.md`), replace the entire guardrails section
with the correct variant. Report:
`✅ <filename>: guardrails corrected`

**Rationale**: The current idempotency logic skips files that
already have a `## Guardrails` heading, which means running
`/uf.init` after this fix would NOT correct existing incorrect
guardrails. The replace-if-incorrect behavior ensures the fix
is applied to all repos on next `/uf.init` run, not just new
scaffolds.

**Implementation**: For each execution/utility command, define
a "correctness marker" — a unique phrase that MUST appear in
the correct guardrail. If `## Guardrails` exists but the
correctness marker is absent, replace the section.

| Command | Correctness marker |
|---------|-------------------|
| `speckit.implement.md` | "writes source code" |
| `speckit.constitution.md` | ".specify/memory/" |
| `speckit.taskstoissues.md` | "GitHub issues via" |

### D3: Remove the workaround Note

**Decision**: Delete the Note at current lines 627–632 that
documents the implement override conflict.

**Rationale**: The Note documents a workaround for a problem
that the fix eliminates. Keeping it would be zero-waste
violation — documenting a conflict that no longer exists.

### D4: Guardrail content regression tests

**Decision**: Add tests that verify the guardrail templates
in `uf.init.md` contain correct content for each command type.

**Rationale**: The triage panel unanimously identified zero
existing guardrail content validation tests. The existing
regression test pattern (`TestScaffoldOutput_No*References`)
provides a natural home for these assertions.

**Test assertions**:
- Spec-phase variant MUST contain "NEVER modify source code"
- Implement variant MUST NOT contain "NEVER modify source code"
- Implement variant MUST contain "writes source code"
- Constitution variant MUST contain ".specify/memory/"
- Constitution variant MUST NOT restrict writes to FEATURE_DIR
- Taskstoissues variant MUST contain "GitHub issues"
- Taskstoissues variant MUST contain "matching the current
  Git remote" (repository-scoping constraint)
- Taskstoissues variant MUST NOT contain "files this command
  may write"

## Risks / Trade-offs

### R1: Existing repos need a re-run

Repos that already have incorrect guardrails injected will
retain them until `/uf.init` is re-run. Decision D2
(replace-if-incorrect) mitigates this — the next `/uf.init`
run will correct the guardrails automatically.

### R2: Three more template blocks increase Step 6 length

Step 6 grows from ~90 lines to ~140 lines. This is acceptable
because the content is now correct — shorter-but-wrong is not
a valid trade-off against longer-but-right.

### R3: speckit.testreview.md not addressed

The 4th potentially-incorrect guardrail (`speckit.testreview.md`
claims file write permissions but is read-only) is deferred.
The error is harmless since the command's own `STRICTLY READ-ONLY`
constraint is stricter. Including it would expand scope beyond
the 3 cases identified in #256. To be addressed as part of
the preset migration (#257) or a standalone cleanup chore.

### R4: Replace-if-incorrect may overwrite user customizations

If a downstream repo has manually customized the guardrails
section (added constraints, reworded content), the
replace-if-incorrect logic (D2) would overwrite those
customizations when the correctness marker is absent.
Mitigation: guardrails are generated by `/uf.init`, not
hand-authored; files are under VCS for recovery; the
`✅ guardrails corrected` report line makes replacement
observable.
