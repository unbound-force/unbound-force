# Key Technical Decisions

## Two diff files: filtered vs annotated

**Decision**: Produce both `pr-diff-filtered.patch` (standard unified diff) and `pr-diff-annotated.patch` (with `[L<N>]` line prefixes).

**Why**: LLMs consistently confuse patch-file positions (sequential line count across the whole multi-file patch) with source-file line numbers (from `@@` hunk headers). For example, a 208-line file was getting comments on "line 243" because the model counted from the top of the concatenated patch. The `[L<N>]` prefix gives the model the correct number to read directly.

**Why two files**: `filter-diff-lines.py` needs standard unified diff format to parse `@@` hunk headers for validation. Annotated prefixes would break its parser. So the un-annotated version is used for validation, and the annotated version goes to the LLM.

**Pipeline**: `raw → filtered (noise removed) → annotated (line-numbered) → LLM`

## Three-workflow chain for fork PR support

**Decision**: Use collect + consumer + reusable instead of a single workflow.

**Why**: Fork PRs cannot access the base repo's secrets. The `pull_request` event fires in the fork's context. A `workflow_run` or `workflow_dispatch` trigger in the base repo runs with the base repo's secrets. The collect workflow (fork-safe, no secrets) uploads artifacts; the consumer downloads and authenticates.

**Alternative considered**: Single workflow with `pull_request_target` — rejected because it runs the workflow definition from `main`, not the PR branch, making it impossible to test workflow changes.

## Comment cleanup: delete + minimize

**Decision**: Delete issue comments and PR review comments, minimize Reviews API objects.

**Why**: The Reviews API (`POST /pulls/{n}/reviews`) creates objects that cannot be deleted via REST. GraphQL `minimizeComment` with classifier `OUTDATED` collapses them. Individual PR comments and issue comments can be deleted via REST.

**Alternative considered**: Supersede reviews with "Superseded by updated council review" body — rejected because it cluttered the PR timeline.

## continue-on-error on the review step

**Decision**: `continue-on-error: true` on the "Run council review" step in the caller workflow (`reusable_council_review.yml` in org-infra), not in this composite action.

**Why**: The council review is supplemental — a failure should not block CI. The downstream posting steps check `steps.review.outcome == 'success'` and skip if the review failed. This is the caller's responsibility, not the action's.

**Known gap**: No failure notification step yet. When the review fails silently, maintainers get no signal. A future improvement should add a step that posts a notice annotation on `outcome == 'failure'`.

## SHA-pinned action references

**Decision**: Pin the `council-review-action` to a full commit SHA, not a branch or tag.

**Why**: SHA-pinning is a supply-chain security best practice (scored by OSSF Scorecard). The SHA is immutable — even if the branch is rebased or force-pushed, the pinned commit stays the same.

**Current state**: Pinned to the `feat/council-review-action` feature branch SHA. Will be updated to a `main` SHA after the PR merges.

## Noise file filtering

**Decision**: Exclude lock files, vendored deps, generated code, test fixtures from the diff before review.

**Why**: These files add noise without review value. They inflate token costs and dilute review quality.

**Explicitly NOT excluded**: Spec files (openspec/, .specify/, docs/) — Divisor personas specifically review specs for intent drift and completeness.

## Prompt injection defense

**Decision**: Treat all diff content, PR titles, and file content as untrusted input. The prompt explicitly instructs the model to ignore override attempts.

**Scope**: This is a defense-in-depth measure. The primary defense is that the model runs in a sandboxed CI environment with no shell access, no network access beyond Vertex AI, and no write permissions beyond the review JSON output.

## OpenCode version pinning

**Decision**: Pin `opencode-ai@1.2.26` rather than using latest.

**Why**: Newer versions (1.17.x) introduced breaking changes in the `run` command output format. Version 1.2.26 is validated to work with the JSONL parsing pipeline.

## Pre-fetched context (full bodies)

**Decision**: Keep full review bodies, inline comments, and issue bodies in the pre-fetched context — no truncation.

**Why**: Human reviewer feedback is high-signal context. Truncating it causes the AI to duplicate findings or contradict prior feedback. The total token cost (~25-30K tokens) is well within the 200K context window.
