## Why

Issue #570 identifies a contradiction: the issue title states "Dewey learnings are supposed to be part of the PR," but a cited review comment flagged learnings as "noise" to strip from the feature PR. The current `AGENTS.md` **Commit scope** rule (lines 192-198) says "Only commit files directly related to the active spec or change" — which reviewers may misinterpret as excluding `.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md` files produced during the retrospective (Step 9 of `/uf.unleash`).

Without explicit policy, agents and reviewers lack authoritative guidance on whether learnings ride along in the feature PR or belong on a separate branch. This ambiguity produces inconsistent behavior: some PRs include learnings, others strip them, and issue #570 remains unresolved.

Closes #570.

## What Changes

Documentation-only policy codification in `AGENTS.md` and the `/uf.unleash` Step 9 (Retrospective) command:

1. **`AGENTS.md` — Commit scope rule**: Add an explicit clarification that Dewey learnings (`.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md`) produced during the change's workflow ARE directly related to the active spec or change and MUST be included in the feature PR.
2. **`AGENTS.md` — Documentation gate**: Add a note that learnings produced by the retrospective satisfy the documentation gate — they are intentional documentation artifacts, not unrelated scope noise.
3. **`.opencode/commands/uf.unleash.md` — Step 9 (Retrospective)**: Add a policy note stating that learnings stored via `dewey_store_learning` produce files that are part of the feature PR scope.

No code changes. No changes to `.gitignore`. No changes to retrospective mechanics or how learnings are produced.

## Capabilities

### New Capabilities
- `learnings-in-pr-policy`: Explicit documentation stating that Dewey learnings produced during a change's workflow are intentional parts of the feature PR, not unrelated scope noise.

### Modified Capabilities
- `commit-scope-rule`: Clarified to explicitly include `.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md` as in-scope for the feature PR.
- `documentation-gate`: Clarified to note that retrospective learnings satisfy the documentation gate as intentional artifacts.
- `uf-unleash-step-9`: Annotated with a policy note that learnings are PR-scoped.

### Removed Capabilities

None.

## Impact

- **Affected files**: `AGENTS.md` (behavioral rules section), `.opencode/commands/uf.unleash.md` (Step 9 section).
- **Affected workflows**: All OpenSpec and Speckit changes that produce learnings via `/uf.unleash` Step 9. Reviewers will have explicit policy to reference when evaluating whether learnings are in-scope.
- **No behavioral change**: The retrospective mechanics, learning production, and `.gitignore` remain unchanged. This is purely a documentation clarification.
- **Cross-repo**: Hero repositories that follow the same commit-scope pattern SHOULD align their own documentation, but this change is scoped to the meta-repo only.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Learnings are self-describing artifacts: each learning file includes producer identity, timestamp, tags, and category metadata. Codifying that they ride along in the feature PR makes them discoverable by other heroes (e.g., Mx F can find retrospective learnings in the same PR that produced them). This strengthens artifact-based communication — the learning is co-located with the work that generated it, not stranded in a separate branch that requires explicit coordination to find.

### II. Composability First

**Assessment**: PASS

This change introduces no new dependencies. Learnings are standalone markdown files that provide value independently (future semantic search, cross-session context). Codifying their PR inclusion does not require any hero to be present — it simply clarifies where the files live in the existing workflow.

### III. Observable Quality

**Assessment**: PASS

This change directly supports traceability — a core tenet of Observable Quality. Learnings with provenance metadata (producer, branch, date, category) are more traceable when they are included in the PR that generated them: the PR diff shows what was learned, the commit history links learning to work, and reviewers can verify the learning's relevance. Stripping learnings from the PR breaks this traceability chain.

### IV. Testability

**Assessment**: N/A

This is a documentation-only change with no code components. No testable side effects.
<!-- scaffolded by uf vdev -->
