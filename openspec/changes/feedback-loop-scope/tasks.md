## 1. Address feedback

- [x] 1.1 Add severity to the item fields and the triage display.
- [x] 1.2 Add the severity scope to Section 2.5 and keep MEDIUM and
  LOW items at Tier 1.
- [x] 1.3 Add the MEDIUM and LOW option set and the FILE and REPLY
  decisions, replies, issue filing with its failure path, and thread
  resolution.
- [x] 1.4 Replace the review-council gate with the pre-flight skill
  in hard-gate mode and rename the `review-council` progress step to
  `verify`.
- [x] 1.5 Keep learnings out of feedback commits.
- [x] 1.6 Document the `filed:` and `reply:` prefixes in
  `schemas/feedback-triage/README.md`.
- [x] 1.7 Checkpoint: `go test ./internal/scaffold/` passes.

## 2. Review PR

- [x] 2.1 Add `decideReviewMode` and the `resolve_review_mode` tool
  to the review-dispatch plugin, with the thresholds as constants.
- [x] 2.2 Add Step 3.11 (calls the tool) and Step 3.12 Light Check
  with untrusted-input framing.
- [x] 2.3 Checkpoint: `resolve-review-mode.test.ts` and the tool
  inventory tests pass.

## 3. Learnings

- [x] 3.1 Exclude `.uf/dewey/learnings` in `uf_finale_stage`.
- [x] 3.2 Update `uf.finale.md`, `uf.unleash.md`, and `AGENTS.md`.
- [x] 3.3 Move `dewey-learning-pr-policy` and `finale-dewey-commit`
  to the archive without merging their specs.
- [x] 3.4 Checkpoint: `uf-finale-stage-git.test.ts` passes against
  real git.

## 4. Documentation and verification

- [x] 4.1 Mirror every command and plugin change in the scaffold
  assets.
- [x] 4.2 Update `docs/usage.md` and `docs/architecture.md`; file the
  documentation issue (#996) and the tutorial issue (#997).
- [x] 4.3 Pin the new rules in `command_contract_test.go` and update
  the option-label pins and `uf_finale_stage` tests.
- [x] 4.4 Run `go test ./...` and `make plugin-test`.
- [x] 4.5 Run `openspec validate feedback-loop-scope --strict`.
