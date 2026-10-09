<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. OpenSpec Schema and CLI

- [x] 1.1 N/A — `openspec` CLI is an external tool; the
  field is added directly to `.openspec.yaml` (done:
  this change's `.openspec.yaml` includes
  `originating_issue: 554`).
- [x] 1.2 N/A — `openspec` CLI is external; flag belongs
  in the upstream tool.
- [x] 1.3 Updated the OpenSpec proposal template
  (`internal/scaffold/assets/openspec/schemas/unbound-force/templates/proposal.md`)
  with a comment documenting the optional field.
- [x] 1.4 N/A — `openspec` CLI tests are external.

## 2. Speckit Template

- [x] 2.1 Updated the OpenSpec proposal template (the
  equivalent template in this repo) with a comment
  explaining the `originating_issue` field purpose.
- [x] 2.2 N/A — `specify init` is an external tool.

## 3. /uf.finale PR Body Generation

- [x] 3.1 Updated `.opencode/commands/uf.finale.md` Step 5d
  to detect the change directory from the branch
  name (`opsx/<name>` → `openspec/changes/<name>/`,
  `NNN-*` → Speckit spec directory).
- [x] 3.2 Added logic to read `originating_issue` from
  `.openspec.yaml` (OpenSpec) or `proposal.md`
  frontmatter (Speckit).
- [x] 3.3 When `originating_issue` is present, append
  `Closes #<N>` immediately after the `## Summary`
  section in the generated PR body. When absent,
  emit no `Closes` line.
- [x] 3.4 Added a soft warning when on an `opsx/*` branch
  with no `originating_issue` set, suggesting the
  developer add it manually to `.openspec.yaml`.

## 4. Tests

- [x] 4.1 N/A — `/uf.finale` is a markdown command file,
  not Go code; scaffold drift tests cover sync.
- [x] 4.2 N/A — same as 4.1.
- [x] 4.3 N/A — `openspec` CLI is external.
- [x] 4.4 N/A — `openspec` CLI is external.

## 5. Documentation

- [x] 5.1 Updated `AGENTS.md` to document the
  `originating_issue` field.
- [x] 5.2 Updated `CHANGELOG.md` with an entry for this
  change.

## 6. Constitution Alignment Verification

- [x] 6.1 Verified constitution alignment:
  - I. Autonomous Collaboration: PASS — field is in
    artifact metadata (`.openspec.yaml`), no runtime
    coupling.
  - II. Composability First: PASS — field is optional,
    backward compatible.
  - III. Observable Quality: PASS — field is
    machine-parseable YAML integer.
  - IV. Testability: PASS — behavior is deterministic
    and verifiable via scaffold drift tests.

<!-- spec-review: passed -->
<!-- code-review: passed -->
