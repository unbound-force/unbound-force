## Context

The `/uf.unleash` pipeline Step 9 (Retrospective) stores learnings via `dewey_store_learning`, which produces markdown files under `.uf/dewey/learnings/` and `.uf/dewey/compiled/`. These files are committed to the repository as part of the feature branch. However, `AGENTS.md` Commit scope rule states "Only commit files directly related to the active spec or change" — language that reviewers have interpreted as excluding learnings from the feature PR.

Issue #570 surfaced this contradiction: the issue title asserts learnings belong in the PR, but a cited review comment treated them as noise. No governing documentation resolves the ambiguity.

## Goals / Non-Goals

### Goals
- Add explicit policy language to `AGENTS.md` Commit scope rule stating that Dewey learnings produced during a change's workflow are in-scope for the feature PR.
- Add a clarifying note to the Documentation gate section that learnings satisfy the gate as intentional artifacts.
- Annotate `/uf.unleash` Step 9 with a policy note that learnings are PR-scoped.
- Reference issue #570 as the originating context for the policy codification.

### Non-Goals
- Changing how learnings are produced (retrospective mechanics remain unchanged).
- Modifying `.gitignore` or any file-exclusion configuration.
- Changing the `dewey_store_learning` API or Dewey's storage format.
- Adding an `originating_issue` field to the OpenSpec schema (the schema does not currently support this field; fallback is `Closes #570` in the proposal and commit message).
- Modifying any hero repository documentation (scoped to this meta-repo only).

## Decisions

**D1 — Single-rule clarification over new section.** Add the learnings policy as a clarifying clause within the existing Commit scope rule in `AGENTS.md` rather than creating a new standalone rule. Rationale: learnings-in-PR is a scope question, not a new behavioral obligation. Embedding it in the existing rule keeps the behavioral rules compact and ensures readers encounter it in context.

**D2 — Documentation gate note, not a new gate.** Add a brief note to the Documentation gate section rather than creating a third Documentation gate bullet. Rationale: learnings already satisfy the gate's intent (assessing documentation impact). The note makes this explicit without adding procedural overhead.

**D3 — Step 9 annotation, not a workflow change.** Add a `> POLICY NOTE:` blockquote to Step 9 in `uf.unleash.md` rather than restructuring the step. Rationale: the retrospective mechanics do not change. The annotation simply informs agents that the files they produce are PR-scoped, preventing future ambiguity.

**D4 — RFC 2119 language in delta spec.** The spec uses MUST/SHOULD/MAY language to make the policy machine-checkable and reviewable. The Commit scope rule clarification uses MUST to make learnings inclusion mandatory, matching the strength of the existing rule.

**D5 — Constitution alignment via traceability.** The proposal's Observable Quality PASS assessment is the primary design driver: learnings in the PR diff create a verifiable chain from work to learning, supporting the constitution's provenance requirements.

## Risks / Trade-offs

**R1 — Reviewer pushback on PR size.** Some reviewers may still perceive learnings as increasing PR diff size. Mitigation: the policy note makes the inclusion intentional and reviewable — reviewers can assess learning relevance as part of the review, rather than stripping files before review.

**R2 — Schema limitation for originating_issue.** The OpenSpec schema does not support an `originating_issue` field. Mitigation: `Closes #570` in the proposal and eventual commit message provides traceability. If the schema gains this field later, the change can be retroactively annotated.
<!-- scaffolded by uf vdev -->
