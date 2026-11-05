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

## 1. Delegation tool contract

- [x] 1.1 Extend the `uf_delegate_command` input contract
  (.opencode/plugins/uf-workflow/index.ts): add
  `task_classification` (enum `deterministic` | `judgment`,
  optional), `skill` (string, min 1, max 256, optional), and
  `skill_absent_reason` (string, min 1, max 512, optional) to
  `DelegateCommandInputSchema` with superRefine rules —
  `skill` and `skill_absent_reason` mutually exclusive;
  each valid only with `deterministic` classification
  (judgment-classified and unclassified command-path calls
  supplying either field are rejected); prompt-only delegation
  (no `command_path`) requires `task_classification`. Update the registered
  `args` descriptions to match the schema (registration
  parity). Unit tests for each validation rule, including
  the prompt-only classification requirement, the command-path
  exemption, over-length and empty-string rejection, and
  judgment-classified `skill` rejection.
- [x] 1.2 Enforce the deterministic reference requirement in
  `execute()`: a `deterministic` classification supplying
  neither `skill` nor `skill_absent_reason` returns a
  non-retryable `SKILL_REFERENCE_REQUIRED` envelope whose
  message names the remediation (pass a skill reference,
  record the absence reason, or classify as judgment).
  Validation runs before host-model resolution and child
  spawn. Envelope tests: failure code, non-retryable flag,
  remediation text, zero session activity.
- [x] 1.3 Implement skill resolution: a `skill` value without
  `/` resolves against `.opencode/skills/<name>/SKILL.md`
  then `.opencode/skill/<name>/SKILL.md` via the injected
  `fileExists` dependency; a value containing `/` resolves
  as a repo-relative path through the existing
  traversal guard; an unresolvable reference returns a
  non-retryable `SKILL_NOT_FOUND` envelope naming the
  searched locations. Tests: both skill roots, same-name
  collision across the roots (first root wins), script-path
  resolution, traversal-rejected path, not-found envelope,
  no child session spawned on failure.
  (Superseded by `retroactive-push-skill-contracts`,
  openspec/changes/retroactive-push-skill-contracts/: the
  legacy `.opencode/skill/` root is retired and migrated by
  `uf init`; a `/`-free name that misses the skill directory
  falls back to a repo-root file reference, and an existing
  directory is not a file reference. That change MUST archive
  after this change.)
- [x] 1.4 Inject the tool-rendered skill directive into the
  composed prompt at a fixed position (after command text
  and feature-dir context, before the extra prompt), wording
  per design D3, instructing the child to load the skill or
  run the script and execute its full procedure. Golden-shape
  test asserting the exact composed prompt with and without
  command text.
- [x] 1.5 Record the audit fields: `checkpointData` and
  `DelegateCommandData` gain `task_classification`, the
  resolved `skill` reference, and `skill_absent_reason` when
  present; `skill` and classification also surface in the
  structured output. Tests: checkpoint data round-trips
  through the progress store; envelope fields
  machine-parseable; callers omitting the new fields
  unchanged (backward compatibility).

## 2. Command template

- [x] 2.1 Rewrite the implement section of
  .opencode/commands/uf.unleash.md: replace the "Mechanical
  tasks use repo tooling" prose block (dropping its
  renumber-adr example, which references tooling from another
  project) with a self-describing classification step (classify
  each task as deterministic or judgment before delegation; no
  citation of repo-local skills — the template is
  scaffold-embedded) and the parameter contract; update every
  `uf_delegate_command` call site in the template,
  grep-verified — the quick-tier whole-change implement
  delegation passes `task_classification` derived from the
  nature of the work (deterministic with `skill` or
  `skill_absent_reason` only when a repo skill or script
  performs the change end-to-end; judgment otherwise); per-task
  implement delegations pass `task_classification` plus
  `skill` or `skill_absent_reason`; code-fix delegations
  declare `task_classification` derived from the nature of the
  fix (deterministic with a verified skill or script reference
  when one performs the fix; judgment for prose-finding
  application); command-path
  Steps 2/3 calls remain unchanged; and the Swarm
  parallel-execution spawn prompts carry the classification
  and, for deterministic `[P]` tasks, the verified skill or
  script reference in the subtask description.

## 3. Governance layers

- [x] 3.1 [P] Add DE-008 [MUST] to
  .opencode/uf/packs/determinism.md in a new Delegation Skill
  Enforcement section: deterministic-classified delegation
  MUST carry a verified skill/script reference or a
  substantive recorded absence reason (stating what was
  searched for or what was found); delegation instruction
  MUST use the
  parameter contract, not prose-only skill instructions.
  Bump the pack version 1.2.0 → 1.3.0 and add DE-008 to the
  severity-mapping MUST/CRITICAL list.
- [x] 3.2 [P] Update .opencode/skills/determinism-first/
  SKILL.md: body gains the delegation rule (declare
  `task_classification`; pass a verified `skill` or record
  `skill_absent_reason`); the Related Enforcement section
  lists DE-008 and the enforcement parameter contract
  alongside the existing entries.
- [x] 3.3 [P] Add a Deterministic Delegation Discipline
  section to the Code Review Mode checklist of
  .opencode/agents/divisor-guard.md — self-describing (names
  the parameter contract; no DE-008 citation, since the agent
  definition is scaffold-embedded and the pack is repo-local):
  flag delegation instruction binding mechanical work to skills
  via prose only, deterministic delegations lacking a
  skill reference or recorded absence reason (checkpoint
  data is the audit source), vacuous absence reasons (a
  recorded reason that does not state what was searched for
  or what was found), and Swarm-path spawn prompts for
  deterministic `[P]` tasks that omit the classification or
  skill reference.

## 4. Tests and seam verification

- [x] 4.1 Add spec-scenario coverage to
  .opencode/test/unleash-delegate.test.ts mirroring the
  unleash spec delta: deterministic delegation with skill
  (resolved, injected, recorded); deterministic without
  skill or reason (SKILL_REFERENCE_REQUIRED); unresolvable
  reference (SKILL_NOT_FOUND); recorded absence reason
  proceeds; judgment pass-through; prompt-only without
  classification rejected; over-length, empty-string, non-enum
  classification values, and judgment-classified `skill`
  rejected; quick-tier whole-change classification
  pass-through; skill-root collision resolving to the first
  root; script-path
  skill; command-path
  delegation unaffected.
- [x] 4.2 DE-007 seam verification naming the three layers:
  registration surface (the registered `args` match the
  documented contract — update the registration assertions
  in .opencode/test/uf-workflow.test.ts and
  .opencode/test/scratch-smoke.test.ts if they pin the
  argument set); documented output shape (golden assertions
  on the envelope and checkpoint fields); write-then-read
  round-trips (checkpoint data recorded by a delegated run
  readable through the progress store).

## 5. Scaffold sync and verification

- [x] 5.1 Update the embedded mirrors for every changed
  source — internal/scaffold/assets/opencode/plugins/
  uf-workflow/index.ts, .../commands/uf.unleash.md, and
  .../agents/divisor-guard.md — and keep the
  drift-detection tests green (the determinism pack and
  determinism-first skill are repo-local,
  knownNonEmbeddedFiles, and need no mirrors).
- [x] 5.2 CI-parity pre-flight: `make check` (lint, test,
  build, coverage-gate) and `make plugin-test`; all green
  before any task in this change is marked complete.
- [x] 5.3 Constitution-alignment verification (each principle
  re-checked against the final diff; DE-008 strengthens the
  determinism gate — no severity, threshold, or existing rule
  weakened) and the documentation gate with per-file
  dispositions: a CHANGELOG entry per the spec's
  change-record requirement plus an assessed disposition for
  each of docs/usage.md, docs/architecture.md,
  docs/heroes.md, AGENTS.md, and README.md (updated, not
  applicable, or covered by the filed documentation issues).
- [x] 5.4 Before filing, check open issues in this repo and in
  unbound-force/website for existing coverage of this change's
  documentation needs, then file the mandatory documentation
  issues for this user-facing change (it modifies the
  `/uf.unleash` workflow contract and the Guard's review
  capabilities) and record both filed issue numbers in the
  change record before PR merge: a docs-labeled issue in this
  repo (AGENTS.md documentation gate) and a GitHub issue in
  unbound-force/website (constitution Cross-Repo Documentation
  rule). At filing time, reassess the blog/tutorial
  disposition recorded in the proposal's Impact section and
  record the outcome with the issue numbers.
- [x] 5.5 File a Replicator-side issue in
  unbound-force/replicator for forge spawn-subtask
  parameter-contract parity (task classification plus
  skill/script reference parameters on the subtask spawn
  tool, mirroring `uf_delegate_command`'s contract), linking
  this change, and record its number in the change record —
  tool-level validation on the Swarm path is Replicator-owned
  (design Non-Goal) and this issue is its tracked closure.

<!-- scaffolded by uf vdev -->
