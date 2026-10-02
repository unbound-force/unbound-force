## Context

`/uf.review-pr` delegates PR analysis to a sub-agent that auto-loads AGENTS.md. AGENTS.md's gatekeeping rules (Behavioral Rules > Gatekeeping) prohibit agents from modifying quality/governance gate values, and the rules text itself contains the marker strings `<!-- code-review: passed -->` and `<!-- spec-review: passed -->` as examples of protected artifacts. When a PR diff includes an OpenSpec or Speckit `tasks.md` file that legitimately contains these markers (written by `/uf.review-council` during `/uf.unleash`), the sub-agent flags them as gatekeeping violations because it cannot distinguish a legitimately-earned marker from one that was illegitimately added via a static diff.

The markers are written at `uf.unleash.md` lines 572-595 (code review) and earlier in the spec review phase. They appear in `openspec/changes/*/tasks.md` and `specs/*/tasks.md` files. They are a resumability record, not a gate weakening.

## Goals / Non-Goals

### Goals

- Eliminate the false positive: `/uf.review-pr` MUST NOT flag `<!-- code-review: passed -->` or `<!-- spec-review: passed -->` markers in OpenSpec/Speckit task files as gatekeeping violations.
- Preserve true positive detection: `/uf.review-pr` MUST continue to flag real weakening of coverage/CI/severity/convention/constitution gates.
- Provide a regression test that reproduces the false positive scenario and asserts it is no longer reported.
- Keep the exception narrow and precisely scoped to the exact marker strings and file paths.

### Non-Goals

- Modifying AGENTS.md gatekeeping rules themselves.
- Changing `/uf.review-council` or `/uf.unleash` marker-writing logic.
- Expanding into acceptance-criteria-review work (#563).
- Adding a general-purpose "allow-list" mechanism for gate markers.
- Changing the sub-agent's auto-loading of AGENTS.md.

## Decisions

### D1: Exception location — sub-agent prompt pre-filter

**Decision**: Add the exception as a dedicated clause in the sub-agent prompt (Step F.3 Constitution Compliance, or a new pre-filter step before it), not as a post-hoc filter on findings.

**Rationale**: Catching the exception before the AI review step prevents the sub-agent from even generating the false-positive finding, reducing token waste and avoiding the need to redact findings after the fact. It also makes the exception visible to anyone reading the prompt, improving maintainability. Placing it in Step F.3 specifically (rather than a new step) keeps the step count stable and co-locates it with the constitution compliance logic that produces the finding.

### D2: Exception scope — exact strings + exact paths

**Decision**: The exception MUST match ALL of the following:
- Marker string: `<!-- code-review: passed -->` OR `<!-- spec-review: passed -->` (exact, case-sensitive).
- File path: matches `openspec/changes/*/tasks.md` OR `specs/*/tasks.md` (relative to repo root).

**Rationale**: Narrow scoping prevents the exception from being exploited to hide illegitimate gate modifications. A marker in any other file path (e.g., `AGENTS.md`, `README.md`, arbitrary `.md` files) MUST still be flagged. A different marker string in a task file MUST still be flagged. Both conditions must be true simultaneously.

### D3: Exception mechanism — prompt instruction, not code

**Decision**: The exception is implemented as a natural-language instruction in the sub-agent prompt, not as a code-level filter in a Go binary or script.

**Rationale**: `/uf.review-pr` is an OpenCode command (a markdown prompt), not a compiled binary. The sub-agent is an LLM that follows prompt instructions. Adding a code-level filter would require a fundamental architectural change to how commands work. Prompt instructions are the native mechanism for constraining sub-agent behavior in this system.

### D4: Regression test — pipeline test in council-review-action

**Decision**: The regression test lives in `council-review-action/test/` (or the appropriate test directory for the review pipeline), feeding a synthetic diff containing legitimate markers in a `tasks.md` file and asserting no gatekeeping finding is produced.

**Rationale**: The review pipeline already has 91 assertions in its test suite. Adding the regression test there follows existing conventions and ensures the false positive cannot recur without CI catching it. The test is isolated and does not require external services (constitution principle IV).

### D5: No changes to AGENTS.md or constitution

**Decision**: AGENTS.md gatekeeping rules and the org constitution remain unchanged.

**Rationale**: The gatekeeping rules are correct — agents MUST NOT modify gate values. The problem is a false positive in the detection mechanism, not a flaw in the rule itself. Changing the rule to accommodate the false positive would weaken the gate for all other scenarios.

## Risks / Trade-offs

### R1: Prompt instruction reliability

**Risk**: LLM sub-agents may not perfectly follow the exception instruction, especially under context compression or with adversarial diffs.

**Mitigation**: The regression test catches regressions. The exception is narrow and unambiguous (exact strings + exact paths), making it easier for the LLM to apply correctly. If false positives persist, the exception can be elevated to a code-level pre-filter in a future change.

**Accepted**: Yes — the current false positive is 100% reliable (it always fires), so even imperfect prompt compliance is an improvement.

### R2: Exception could be exploited

**Risk**: An attacker could place a legitimate-looking marker in a `tasks.md` file to hide a real gate modification.

**Mitigation**: The exception only suppresses the gatekeeping finding for the exact marker strings in task files. Other findings (security, alignment, CI failures) are unaffected. The markers themselves are also verified by the `/uf.review-council` during `/uf.unleash` — a PR that adds markers without having run the pipeline would fail other checks (e.g., the markers would not match the actual review state).

**Accepted**: Yes — the risk is low because the exception is narrow and the markers are corroborated by pipeline state.

### R3: Scope creep into #563

**Risk**: The change could expand into acceptance-criteria-review work.

**Mitigation**: The proposal, design, and specs explicitly exclude #563 scope. The tasks.md will have a single task group focused on the marker exception. The regression test is scoped to the marker false positive only.

**Accepted**: Yes — explicit non-goal statement and narrow task scope prevent creep.

<!-- scaffolded by uf vdev -->
