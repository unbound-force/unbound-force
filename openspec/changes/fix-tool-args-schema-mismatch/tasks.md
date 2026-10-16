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

## 1. Canonical plugin conversions (different files)

- [x] 1.1 [P] In `.opencode/plugins/uf-workflow/index.ts`,
  rename `parameters:` to `args:` in the eight factories and
  convert every plain `{ type, description }` object to zod
  shapes preserving name, type, optionality, and description
  exactly (no `.default()`, execute bodies untouched):
  `createCheckpointTool` (`command`, `step` required;
  `data`, `key` optional strings), `createGetProgressTool`
  and `createClearProgressTool` (`command` required, `key`
  optional), `createNormalizeModelNameTool` (`rawModelId`
  required), `createGitFetchTool` (`remote` required,
  `branch` optional), `createGitPushTool` (`setUpstream`
  optional boolean, `branch` optional string),
  `createGitMergeTool` and `createGitRebaseTool` (`ref`
  required). Follow the `createPersistUnleashSummaryTool`
  declaration style (`z.string().describe(...)`,
  `z.boolean()`, `z.number()`)
- [x] 1.2 [P] In `.opencode/plugins/uf-finale/index.ts`, add
  `import { z } from "zod"` (after the `@opencode-ai/plugin`
  import, matching uf-workflow's grouping) and convert the
  three factories with the same fidelity rules:
  `createUfFinaleCommitContextTool` (`rawModelId` required
  string), `createUfFinaleCreatePrTool` (`title`, `body`
  required; `targetRepo` optional string),
  `createUfFinaleCiPollTool` (`prNumber` required number,
  `timeoutMinutes` optional number — the `?? 15` default
  stays inside `execute`)

## 2. Schema-exposure regression tests (different files)

- [x] 2.1 [P] Add schema-exposure tests to
  `.opencode/test/uf-workflow.test.ts` following the
  `invoke-agent.test.ts` "exposes ... tool definition"
  precedent: for each of the nine factories (the eight
  converted plus `createPersistUnleashSummaryTool` as a
  future-regression guard), construct the tool with injectable
  dependencies and assert
  `Object.keys(definition.args).sort()` equals the expected
  sorted key list; assertions must fail if `args` is
  undefined or a parameter is dropped or renamed
- [x] 2.2 [P] Add schema-exposure tests to
  `.opencode/test/uf-finale.test.ts` for the three converted
  factories, same pattern and failure semantics

## 3. Scaffold mirror sync (after canonical files final)

- [x] 3.1 [P] Mirror the uf-workflow plugin:
  `cp .opencode/plugins/uf-workflow/index.ts
  internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
- [x] 3.2 [P] Mirror the uf-finale plugin:
  `cp .opencode/plugins/uf-finale/index.ts
  internal/scaffold/assets/opencode/plugins/uf-finale/index.ts`

## 4. Documentation

- [x] 4.1 Add a `fix-tool-args-schema-mismatch` entry under
  `## Unreleased` → `### Fixed` in `CHANGELOG.md` (create the
  section if absent) following the
  `- <change-name>: <summary>` format with
  `(Spec: openspec/changes/fix-tool-args-schema-mismatch/,
  Refs: #732)`, summarizing: uf-workflow and uf-finale tool
  factories now declare parameters via the SDK-recognized
  zod `args` contract, restoring model-visible schemas and
  runtime argument passing for 11 tools
- [x] 4.2 Assess the documentation gate: file a documentation
  issue in this repository for the user-facing impact
  (restored model-invoked `git_push`, `uf_finale_ci_poll`,
  `uf_finale_create_pr`, and model-name attribution during
  `/uf.unleash` and `/uf.finale`) and record the issue number
  at check-off, or record the exemption with justification
  — filed #740 (this repo) and unbound-force/website#310
  (cross-repo documentation rule)

## 5. Verification

- [x] 5.1 Run `make plugin-test` — full TypeScript suite
  (unit, integration, smoke, coverage) must pass including
  the new schema-exposure tests and all unmodified
  execute-level tests — 23 files, 621 tests passed, exit 0
- [x] 5.2 Run `go test -race -count=1 ./internal/scaffold/...`
  — embedded-asset drift tests must pass after the mirror
  sync — ok, exit 0
- [x] 5.3 Run `make check` for CI parity (lint, test, build,
  coverage gate, plugin tests) — must pass end to end —
  exit 0
- [x] 5.4 Verify FR-001's no-legacy-keys scenario: confirm
  `parameters:` no longer appears in
  `.opencode/plugins/uf-workflow/index.ts` or
  `.opencode/plugins/uf-finale/index.ts` (and therefore not
  in their scaffold mirrors) — grep count 0 in all four
  files
- [x] 5.5 Verify constitution alignment per the proposal:
  self-describing schema contract restored between model and
  tools (I), no new dependencies and standalone plugins (II),
  declaration contract backed by regression tests (III),
  schema tests run in vitest without services (IV), runtime
  validation retained as defense-in-depth (V), deterministic
  tool paths invocable again instead of bash fallbacks (VI)
  — all six principles verified PASS against the implemented
  change
