## 1. Address feedback

- [x] 1.1 Add severity to the item fields and the triage display.
- [x] 1.2 Add the severity scope to Section 2.5 and keep MEDIUM and
  LOW items at Tier 1.
- [x] 1.3 Add the MEDIUM and LOW option set and the FILE and REPLY
  decisions, replies, issue filing, and thread resolution.
- [x] 1.4 Replace the review-council gate with local verification
  and rename the `review-council` progress step to `verify`.
- [x] 1.5 Keep learnings out of feedback commits.

## 2. Review PR

- [x] 2.1 Add the review id to the Step 3.10-ii inline comment fetch.
- [x] 2.2 Add Step 3.11 Re-review Evaluator and Step 3.12 Light Check.

## 3. Learnings

- [x] 3.1 Exclude `.uf/dewey/learnings` in `uf_finale_stage`.
- [x] 3.2 Update `uf.finale.md`, `uf.unleash.md`, and `AGENTS.md`.

## 4. Verification

- [x] 4.1 Mirror every command and plugin change in the scaffold
  assets.
- [x] 4.2 Update the option-label pins and `uf_finale_stage` tests.
- [x] 4.3 Run `go test ./internal/scaffold/` and `make plugin-test`.
- [x] 4.4 Run `openspec validate feedback-loop-scope --strict`.
