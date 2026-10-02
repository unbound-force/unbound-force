## Why

GitHub issue #539 reports a false positive in `/uf.review-pr`: when a PR diff includes an OpenSpec or Speckit `tasks.md` file, the review sub-agent flags the legitimate workflow-gate markers `<!-- code-review: passed -->` and `<!-- spec-review: passed -->` as violations of the AGENTS.md gatekeeping rules. These markers are written by the local `/uf.review-council` during `/uf.unleash` (see `uf.unleash.md` lines 34-46 and 572-595) and serve as the resumability record for the pipeline. They do NOT weaken any gate. The sub-agent auto-loads AGENTS.md, whose gatekeeping rules protect those exact marker strings, and a static PR diff cannot distinguish a legitimately-earned marker from one that was illegitimately added. This causes noise, erodes trust in the review, and blocks PRs that are actually correct.

## What Changes

Add a precise, narrow exception to `/uf.review-pr` so the sub-agent recognizes the legitimate workflow-gate markers (`<!-- code-review: passed -->` and `<!-- spec-review: passed -->`) in OpenSpec/Speckit task files (`openspec/changes/*/tasks.md` and `specs/*/tasks.md`) and does not flag them as gatekeeping violations. The exception MUST:

1. Be scoped to the exact marker strings and the exact file paths where they are legitimately written.
2. Continue to flag any real weakening of coverage/CI/severity/convention/constitution gates.
3. Include a regression test that reproduces the false positive and asserts it is no longer reported.

This change is intentionally SEPARATE from the acceptance-criteria-review work (#563) and must not expand into it.

## Capabilities

### New Capabilities

- `review-pr-marker-exception`: `/uf.review-pr` recognizes legitimate workflow-gate markers in OpenSpec/Speckit task files and suppresses the false-positive gatekeeping finding.

### Modified Capabilities

- `uf.review-pr`: Sub-agent prompt gains a narrow exception clause for `<!-- code-review: passed -->` and `<!-- spec-review: passed -->` markers in `openspec/changes/*/tasks.md` and `specs/*/tasks.md` paths.

### Removed Capabilities

None.

## Impact

- `.opencode/commands/uf.review-pr.md`: Add exception clause to the sub-agent prompt (Step F.3 or a dedicated pre-filter before constitution compliance checks).
- `council-review-action/` or related test fixtures: Add a regression test that feeds a diff containing legitimate markers in a `tasks.md` file and asserts no gatekeeping finding is produced.
- No changes to AGENTS.md gatekeeping rules themselves.
- No changes to `/uf.review-council`, `/uf.unleash`, or the marker-writing logic.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change refines an artifact-based review command. It does not introduce runtime coupling or synchronous interaction. The exception is encoded in the sub-agent prompt (a static artifact), preserving asynchronous, self-describing review output.

### II. Composability First

**Assessment**: PASS

`/uf.review-pr` remains independently usable. The exception is an internal refinement that does not introduce mandatory dependencies on other heroes or tools.

### III. Observable Quality

**Assessment**: PASS

The regression test provides automated, reproducible evidence that the false positive is eliminated. The review output remains machine-parseable and provenance-tracked. The exception narrows findings; it does not suppress provenance or quality metadata.

### IV. Testability

**Assessment**: PASS

A concrete regression test is required as part of this change. The exception logic (file-path + marker-string match) is a pure predicate that is trivially unit-testable in isolation without external services.

<!-- scaffolded by uf vdev -->
