## 1. Tool

- [x] 1.1 Add `renderDeclaredOutput` to
  `.opencode/lib/preflight-declared.ts` (D2-D5) and a
  `rendered_output` field to `DeclaredChecksData`, populated in
  `runDeclaredChecks` and `""` when the config is absent.
- [x] 1.2 Name `rendered_output` in the
  `uf_preflight_declared_checks` description and JSDoc in
  `.opencode/plugins/uf-workflow/index.ts`.

## 2. Skill

- [x] 2.1 In `.opencode/skills/pre-flight/SKILL.md`:
  - add `rendered_output` to the Phase 2b example;
  - replace the "non-PASS `output`" prose with a rule to insert
    `rendered_output` verbatim (omitted when empty);
  - add a "Project-declared check output" section to both Phase 5
    templates;
  - extend the Phase 5 note (no summarizing, re-fencing or
    trimming);
  - record the line delta (589 -> 598).
- [x] 2.2 Mirror the lib, plugin and skill into
  `internal/scaffold/assets/opencode/` byte-identically.

## 3. Tests

- [x] 3.1 [P] `preflight-declared.test.ts`: the exact block format;
  fence safety (a hostile output with a four-backtick line and a
  fake verdict heading gets a five-backtick fence; a single
  backtick gets the minimum three); omission of empty, whitespace
  and skipped checks; the timeout marker; advisory-pass output in
  the chain; `""` for an absent config; and `rendered_output` in
  the determinism payload.
- [x] 3.2 [P] `golden-output-shape.test.ts`: pin `rendered_output`
  in the success case.
- [x] 3.3 [P] `preflight-declared-roundtrip.test.ts`: a real
  advisory script whose output contains a triple-backtick line is
  rendered with a four-backtick fence, and a silent check is
  omitted.

## 4. Documentation and validation

- [x] 4.1 Add a `CHANGELOG.md` Unreleased -> Added entry.
- [x] 4.2 Capture a Dewey learning in `.uf/dewey/learnings/`.
- [x] 4.3 Run the plugin suite with coverage (floor 94%) and
  `go test ./internal/scaffold/... ./cmd/...` (drift pairs), and
  `openspec validate preflight-declared-output --strict`.
- [x] 4.4 File the documentation issues required by the
  documentation gate: one for this repository (the Phase 5
  "Project-declared check output" section and `rendered_output`;
  unbound-force/unbound-force#926 covers only `.uf/preflight.yaml`),
  and a comment extending unbound-force/website#373 to cover
  check-output evidence. Record both in the PR description.
  Filed unbound-force/unbound-force#951; extended
  unbound-force/website#373 with a comment.
