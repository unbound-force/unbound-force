<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Tool (#1022)

- [x] 1.1 Extend `.opencode/lib/feedback-replies.ts` (design D1–D7):
  retitle the module header to cover the artifact write; add
  `WriteFeedbackTriageArtifactInputSchema` (1.1.0 item fields plus
  `comment_kind`/`comment_id` join identity, artifact context,
  optional `replies_result` projection, `superRefine` decision
  cross-check); the payload zod mirror of
  `schemas/feedback-triage/v1.1.0.schema.json`; `nextRoundNumber`
  (design D3); `renderFeedbackTriageArtifact` (envelope constants,
  derived summary counts, sorted `divisor_agents_invoked` union,
  identity stripped per `additionalProperties: false`);
  `writeFeedbackTriageArtifact` (join → count → round → render →
  validate → atomic write → clear progress); the optional
  `clearProgress` hook on `FeedbackReplyDeps`; `artifact_error`
  for artifact I/O.
- [x] 1.2 Register `write_feedback_triage_artifact` in
  `uf-workflow`: `createWriteFeedbackTriageArtifactTool` beside the
  other feedback factories with
  `args: WriteFeedbackTriageArtifactInputSchema.shape`, and the
  plugin-entry registration; wire the production `clearProgress`
  through the same branch-keyed state path `clear_progress` uses.
- [x] 1.3 Add `.opencode/test/feedback-replies.test.ts` coverage:
  round scan (first round, gap tolerance, unparsable names ignored,
  read failure), FILE/REPLY routing (join, issue-number and URL
  passthrough, failed-issue nulls, decision-mismatch fail-closed,
  absent `replies_result`), summary counting (decisions, tiers,
  sorted agent union), validation-failure writes nothing, atomic
  write with `0o600`/`0o700` and symlink refusal, clear-progress
  success and warning-on-failure, envelopes, injectable clock.
- [x] 1.4 Pin the payload zod mirror to the checked-in samples
  (design D4): `samples/v1.1.0/sample-feedback-triage.json` and
  the 1.0.0 sample pass; the `invalid-*` samples under
  `samples/v1.1.0/` and `samples/` fail.
- [x] 1.5 Add the tool to the uf-workflow name lists in
  `uf-workflow.test.ts` and `scratch-smoke.test.ts`.
- [x] 1.6 DE-007 seam verification naming the three layers:
  registration surface (the generic `registration-surface.test.ts`
  suite enumerates the tool; `feedback-replies.test.ts` pins
  `args` key parity with
  `WriteFeedbackTriageArtifactInputSchema`); documented output
  shape (`golden-output-shape.test.ts` runs the tool through the
  registry and pins `pr_number`, `path`, `round`, `bytes`,
  `counts`, `progress_cleared`, `warnings`, `summary`);
  write→read round-trips (the written artifact validates under the
  zod mirror and the JSON Schema sample set, reads back from
  disk, and a second call derives the next round from it; a
  cleared progress state reads back through `get_progress` as
  absent). `tool-inventory.ts --seams` reports no missing layer.

## 2. Command and mirrors

- [x] 2.1 Rewrite Section 4.7 of
  `.opencode/commands/uf.address-feedback.md` (design D9): one
  `write_feedback_triage_artifact` call with the parameter mapping
  (items carrying the Section 4.6 comment identity, the
  `post_feedback_replies` success data when it exists), one
  rendered-output example (DE-002), the printed `summary` line,
  and the error-envelope STOP. Keep the no-success Tier 2
  guidance in prose; remove the envelope JSON template, the
  round-scan prose, the atomic-write prose, and the hand-called
  `clear_progress`. Leave Sections 1.2.1, 4.4, 4.5, and 4.6
  untouched.
- [x] 2.2 Update the Progress-Tracking `reply-comments` entry:
  "cleared after Phase 4.7" becomes "cleared by
  `write_feedback_triage_artifact` after the artifact is written".
- [x] 2.3 Copy the command, plugin, and lib to the scaffold
  assets under `internal/scaffold/assets/` byte-identical (no new
  asset-list entries — the lib module is already registered in
  `scaffold.go` and `plugin_activation.go`).
- [x] 2.4 Update `internal/scaffold/command_contract_test.go`:
  re-pin the `feedback triage artifact` clause of
  `TestCommandContracts_AddressFeedback` and the `checkpoint and
  clear` clause of `TestCommandContracts_FeedbackReplyTools` to
  the tool wording; add removed-text checks for the retired 4.7
  prose (the envelope template, the hand-written round scan and
  atomic write).

## 3. Documentation and verification

- [x] 3.1 Add the CHANGELOG entry with the DE-004 instruction
  delta measured by `instruction-lint --base`
  (Section 4.7 before/after lines and tokens).
- [x] 3.2 File the documentation issue for the user-facing change
  (the command's terminal step now reports a tool-rendered
  summary line) in unbound-force/unbound-force before PR merge.
- [x] 3.3 Run `make check` (plugin suite with the 94 percent
  coverage floor, Go drift and contract gates) and
  `make lint-instructions`.
- [x] 3.4 Run `openspec validate extract-feedback-triage-artifact-tool
  --strict`.
- [x] 3.5 Verify constitution alignment as documented in the
  proposal: the artifact envelope and provenance metadata are
  unchanged (I, III), no new dependency or asset-list entry (II),
  every behavior is covered by isolated vitest doubles (IV), the
  write stays inside `.uf/artifacts/feedback-triage/` with
  permission and symlink gates (V), and Section 4.7's mechanical
  prose is gone (VI).
