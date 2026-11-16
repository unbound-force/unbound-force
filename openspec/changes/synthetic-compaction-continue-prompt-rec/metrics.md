# synthetic-compaction-continue-prompt-rec — Change Record

Verification evidence for the change. Task 8.2 (constitution
alignment) and task 8.4 (documentation updates) are recorded
here; task 8.5 (DE-004 instruction-lint delta rows) appends to
this same file when it runs.

**Location rationale**: this change's own artifacts say "the
change record" (proposal.md Impact, tasks 8.2/8.5) without
naming a file, so the repo's established convention governs —
`metrics.md` inside the change directory, the
`command-prose-pilot` precedent (also cited verbatim by
`router-pattern-codification` as "the command-prose-pilot
metrics.md precedent", whose own constitution re-assessment
section this file mirrors).

## Task 8.2 — Constitution alignment (post-implementation)

**Constitution**: `.specify/memory/constitution.md` v1.4.0
(ratified 2026-02-25, last amended 2026-10-14) — **six**
principles. Mapping note, recorded honestly: the task brief
anticipated a four-principle constitution with Security by
Default and Deterministic Execution as sub-facets; the actual
v1.4.0 text declares both as full principles (V at lines
162-197; VI added by the 1.3.1 -> 1.4.0 amendment, lines
199-227). The task's six verification items therefore map 1:1
onto the six declared principles — no remapping needed.

**Verification basis**: the implemented working tree at HEAD
`e180ae79` (the proposal commit) plus the uncommitted phase
2-7 implementation it carries (24 modified files, 5 new test
suites; `git diff --stat` totals +5017/-276). Evidence was
gathered by direct file reads and by read-only test runs on
2026-11-16: `npx vitest run` over the five new suites
(85/85 passed) and the five widened suites (416/416 passed),
and `go test -race -count=1 -run
'TestCommandContracts_SyntheticPromptNotConsent|
TestEmbeddedAssets_MatchSource' ./internal/scaffold/` (ok).
The full `go test ./...` / `make plugin-test` / `make check`
gate is task 8.3 and remains open; nothing here pre-empts it.
Dewey semantic search was unavailable (embedding model not
loaded); Tier 2 keyword search and direct reads were used.

### I. Autonomous Collaboration — PASS

Gate authorization is artifact-mediated, tightening the
principle rather than weakening it:

- A gate override resolves only against the persisted,
  file-backed human-decision record
  `.uf/progress/<key>/human-decisions.json`:
  `resolveHumanDecisionRef`
  (`.opencode/plugins/uf-workflow/index.ts:572-606`) reads it
  through `readHumanDecisionRecord`
  (`.opencode/lib/progress-state.ts:656`). The tool reads
  files, not session context — no runtime coupling.
- Record entries are self-describing artifacts:
  `HumanDecisionEntry` (`.opencode/lib/uf-workflow-types.ts:714`)
  carries `decision_id` (machine-stamped UUID v4), `source`,
  question, answer, and `recorded_at`.
- Accepted overrides carry provenance into the progress
  artifact: the verified reference is stamped as
  `human_decision_ref` in the recorded state (index.ts:843,
  917, 1036), with a caller-supplied key of the same name
  overwritten by the machinery stamp (anti-spoofing; test
  `human-decisions-store.test.ts:1062`).
- The advisory relay renders from the dispatch artifact, never
  child session text: `renderSpecReviewGateRelay`
  (index.ts:1398) + the artifact read at index.ts:1567-1595.
- No synchronous hero interaction is added; every new surface
  (checkpoint validation, two write surfaces, relay predicate)
  completes against local files. The proposal's ex-ante PASS
  assessment holds against the implementation.

### II. Composability First — PASS

- **No new dependency**: the change touches no manifest —
  `git diff --stat -- '*.json'` is empty and `go.mod` is
  absent from the modified-file list. Runtime pins stay
  `@opencode-ai/plugin` 1.19.0 / `zod` 4.6.6.
- **Local enforcement**: every new gate (declared-step
  predicate, override tripwire, escalated-advisory hold) lives
  inside the `checkpoint` tool the `uf.unleash` workflow
  already owns (index.ts:655-1055). No hero, repo, or service
  prerequisite is introduced.
- **Enhanced-when-present, not required**: the
  orchestrator-relay envelope (`human_decisions` on
  `uf_delegate_command`, schema index.ts:3780-3801) activates
  when `/uf.run-pipeline-plan` orchestrates; the standalone
  in-session `record_human_decision` path works without it
  (test `human-decisions-store.test.ts:154` accepts both
  surfaces).
- **Standalone behavior unchanged**: non-override data needs
  no reference (test :1087); plain-APPROVE advancement needs
  no relay (test `spec-review-escalated-advisory.test.ts:394`);
  other commands' checkpoints are unaffected by the
  `uf.unleash` predicate (test `human-decisions-store.test.ts:
  1270`; spec scenario "Another command's checkpoint is
  unaffected").

### III. Observable Quality — PASS

- **Machine-parseable errors**: all four new codes ride the
  existing `failure(message, {retryable, code})` envelope —
  `missing_human_decision` (index.ts:877, 907),
  `undeclared_step` (index.ts:832),
  `escalated_advisory_unrelayed` (index.ts:1024),
  `human_decision_record_unreadable` (index.ts:815, 892,
  1014) — plus `synthetic_prompt_answer` (index.ts:1255) and
  `oversize_step` (index.ts:766). Every rejection names the
  offending step and override-carrying data key with bounded
  echoes (`boundIdentifierEcho`).
- **Golden tests pin the envelopes**:
  `golden-output-shape.test.ts:1893` ("golden output shapes
  -- synthetic-compaction seams (task 7.1)") asserts one
  golden round-trip per new code plus the success envelope
  carrying machine-stamped provenance.
- **Provenance on allowed overrides**: `decision_id`, source,
  and `recorded_at` are machine-stamped at append
  (`appendHumanDecisions`, progress-state.ts:778); the
  advancement checkpoint records which relayed decision
  authorized passage (index.ts:1036).
- **Claims are regression-backed**: the SYNTH-002 prose claim
  is pinned by `TestCommandContracts_SyntheticPromptNotConsent`
  (`internal/scaffold/command_contract_test.go:893`) on both
  live and embedded-mirror copies — verified green in this
  run. Output shapes are stable, zod-schema-declared surfaces
  (`registration-surface.test.ts:308` pins the tool-input
  parity, DE-007 Layer 1).

### IV. Testability — PASS

- **Isolated fixtures, no external services**: every suite
  builds `mkdtemp` roots with an injected `GitExec` branch
  stub and literal timestamps; the incident-replay suite
  states the discipline outright — "no other clocks, no
  network" (`synthetic-compaction-incident-replay.test.ts:
  150`). Conformance to TC-002/TC-005/TC-006 of the
  TypeScript pack (async awaited, per-test setup/teardown in
  `afterEach`).
- **Regression coverage for the incident (SYNTH-005)**: the
  exact `ses_e200afc75ffeSuftvVXtQnBJhR` payload — the
  verbatim 406-character `checklist_override` justification
  and the undeclared step `checklist-override-phase8-9` — is
  embedded as a fixture (incident-replay :83-106, pinned
  verbatim at :205), asserted rejected with nothing persisted
  (:235), and routed under a declared step to observe the
  SYNTH-001 envelope for the same text (:282).
- **Coverage added in phases 2-7, all green 2026-11-16**:
  `human-decisions-store.test.ts` (40 tests: store discipline
  :213/:247/:306/:351, concurrency :385/:410/:428, SYNTH-001
  battery :759-1203, SYNTH-003 battery :1205+),
  `synthetic-compaction-incident-replay.test.ts`,
  `synthetic-compaction-prompt-alone.test.ts`,
  `unleash-step-lockstep.test.ts` (both writer surfaces
  :216/:252),
  `spec-review-escalated-advisory.test.ts` (SYNTH-004 battery
  :275-546), plus the widened `golden-output-shape`,
  `registration-surface`, `uf-workflow`, `unleash-delegate`,
  and `spec-review-gate` suites (416 tests) and the Go
  contract test.
- **Assertions target observable side effects** — returned
  envelopes, persisted state, file modes (e.g. "writes 0o600
  records inside 0o700 directories with no lock or temp
  residue", :213) — not implementation details.
- **Coverage ratchet intact**: the 94% plugin coverage floor
  (`.opencode/vitest.config.ts`) is untouched by this change
  and the new suites run under it (full-suite run is 8.3).

### V. Security by Default — PASS

The synthetic-prompt privilege-escalation path is closed at
three tool-enforced layers, each tested:

1. **Deny-list at recording (SYNTH-006)**:
   `record_human_decision` refuses answers containing the
   normalized synthetic prompt (index.ts:1246-1259, code
   `synthetic_prompt_answer`), including case, whitespace,
   and quoting variation and prompt-embedded-in-larger-text
   (tests :528+).
2. **Override tripwire (SYNTH-001)**: any data key at any
   nesting depth containing "override" requires a
   `human_decision_ref` resolving to a recorded entry
   (index.ts:857-918); free text — including the quoted
   prompt — is rejected at shape check before any record
   lookup (:904), and fabricated references reject
   identically to missing ones (:947).
3. **Declared-step predicate (SYNTH-003)**: the incident's
   undeclared `checklist-override-phase8-9` rejects regardless
   of its data (index.ts:799-844; incident replay :235).

- **Input validation before security-sensitive operations**:
  step names bounded at 128 checked first (index.ts:758-770);
  refs bounded at 256 and UUID v4-shaped before any record
  read (:577-587); question/answer 16 KiB UTF-8 byte bounds
  validated at `safeParse` before path derivation
  (progress-state.ts:778+; tests :351).
- **Store discipline (SYNTH-006)**: `..`-component rejection
  before sanitization (progress-state.ts:616), containment
  guard, symlink refusal on file and directory, 0o600 files in
  0o700 directories, atomic temp-plus-rename writes,
  advisory-lock-serialized appends with stale-lock breaking
  (:704) — tests :213, :247-:350, :385-:428.
- **Fail closed**: corrupted or unreadable records reject with
  `human_decision_record_unreadable` rather than treating the
  failure as authorization (index.ts:882-896; tests :990,
  :1014) — and an absent record is the empty-record case
  reporting `missing_human_decision` (:973), exactly the
  spec's distinction.
- **Bounded-identifier errors only**: no raw question or
  answer text in errors or logs (test :351).
- **VCS exclusion**: the record lives under the gitignored
  `.uf/progress/` tree (.gitignore:51, verified with
  `git check-ignore`).
- **Residual risk, stated honestly**: design.md's threat
  model documents the ceiling — plugin tool factories receive
  only `(exec, cwd)` with no caller identity, so this is not
  a cryptographic boundary against a deliberately adversarial
  session with exec access; it is the same residual risk every
  tool-enforced gate in this suite accepts. Within that
  ceiling the strongest available controls are applied
  (machinery-write relay surface, machine-stamped entries,
  two-deliberate-tool-call authorization, append-only audit
  trail).

### VI. Deterministic Execution — PASS

- **Tool-enforced gate, not prose**: the declared-step
  predicate, override tripwire, and escalated-advisory hold
  execute inside the `checkpoint` tool (index.ts:655-1055)
  with tests — the model cannot reason its way past them
  after compaction, which is exactly the incident mechanism
  this change closes.
- **Deterministic verdict derivation**:
  `isApprovingSpecVerdict` narrowed to plain APPROVE and
  `isEscalatedAdvisorySpecVerdict` added
  (uf-workflow-types.ts:242, :254), so the AWA-to-ESCALATED
  mapping is code, not judgment; the relay predicate checks
  the verdict itself (index.ts:1530-1547).
- **Prescribed output is tool-rendered**:
  `escalatedAdvisoryHoldMessage` (index.ts:629) and
  `renderSpecReviewGateRelay` (index.ts:1398) render the hold
  and relay text deterministically from artifact-sourced data
  — one deterministic message per hold, mirroring the
  HARDEN-002 renderer.
- **Lockstep consistency test**: `unleash-step-lockstep.test.ts`
  enumerates both writer surfaces (command-prose steps and
  plugin auto-checkpoint sites) against
  `UNLEASH_DECLARED_STEP_ALLOWLIST` (index.ts:477), so a newly
  recorded step name missing from the allowlist fails the
  suite instead of wedging the pipeline.
- **Judgment stays in prose at the decision point**: the
  SYNTH-002 not-consent paragraph with its recovery path
  (obtain a real human decision; never compose an entry for a
  decision not received) is the judgment half, pinned by
  contract test — the correct DE split.
- **DE-004 measurable deltas**: instruction-lint rows are
  recorded at task 8.5 (section reserved below); the proposal
  estimates roughly +7 lines x 3 files mirrored byte-identically.

### Gatekeeping rule confirmed

No quality or governance gate was touched: the change
modifies no coverage threshold, CRAP score, severity
definition, CI flag, agent setting, constitution MUST rule,
review limit, or pinned dependency (`vitest.config.ts`,
`coverage-gate.json`, `.github/workflows/`, `go.mod`, and
`.opencode/package.json` are all absent from the diff).
Scaffold mirrors are byte-identical (`cmp` clean on all three
prose pairs plus the lib/plugin mirrors;
`TestEmbeddedAssets_MatchSource` green).

## Task 8.4 — Documentation updates (pre-PR)

Executed 2026-11-16 against the implemented working tree the 8.2
record describes (phases 1-7 and tasks 8.1-8.3 complete; no code
changed for this task). Convention pack: `content.md` (VB-001/002
active voice without weasel words, FA-001 fact verification
before writing, FT-002 code spans, TD-003 no placeholders).

### CHANGELOG.md — entry added

The `synthetic-compaction-continue-prompt-rec` entry was
prepended to `### Fixed` under `## Unreleased` (the section is
newest-first: `harden-spec-review-gate` at commit `4ef06e17`
postdates `dispatch-tier-resolvability` at `703a5229` and sits
above it). It mirrors the `harden-spec-review-gate` entry — the
#1094-era hardening precedent — in shape: incident-summary
opening naming the issue, mechanism sentences carrying
(SYNTH-NNN) refs where that entry carries (HARDEN-NNN), and the
`(Spec: ...)` trailer. It covers all four hardening outcomes
(SYNTH-001 human-decision reference for gate overrides,
SYNTH-002 not-consent prose, SYNTH-003 declared-step rejection,
SYNTH-004 escalated-advisory relay) plus the SYNTH-005 incident
replay. Every claim was verified against the implementation
before writing (FA-001): error codes and tool surfaces in
`.opencode/plugins/uf-workflow/index.ts` (`missing_human_decision`
:877/:907, `undeclared_step` :832, `escalated_advisory_unrelayed`
:1024, `human_decision_record_unreadable` :815/:892/:1014,
`synthetic_prompt_answer` :1255, `record_human_decision`
registered :7927, `human_decisions` envelope :3780), the store
filename at `.opencode/lib/progress-state.ts:570`, the not-consent
prose in all three files, and the fixture payload in
`synthetic-compaction-incident-replay.test.ts:83-106`.

Two deliberate omissions, both forced by implement-time
constraints:

- No `(Docs: #NNNN)` reference — the documentation issue number
  does not exist yet and placeholder text is forbidden (TD-003).
  The orchestrator appends it when filing, the same way
  `harden-spec-review-gate` carries `(Docs: #1091)`.
- No instruction-delta (DE-004) sentence — those numbers are
  task 8.5's deliverable, recorded in its section below when it
  runs.

### No-update rationale — re-verified at implement time, holds

The proposal's Documentation Impact assessment was re-checked
against what was actually implemented (widened uf-workflow plugin
tools, human-decisions store, error envelopes, prose contract
plus scaffold mirrors, new tests). All three no-update rationales
hold:

- **AGENTS.md**: no structural change. `uf-workflow` is already
  the listed plugin (the new tool lives inside it); the new test
  suites live in the already-documented `.opencode/test/`; the
  human-decision record is a gitignored runtime artifact under
  `.uf/progress/`, not a project-structure entry (the same
  treatment `harden-spec-review-gate` gave the progress store it
  introduced); no build command, technology, agent count, or
  convention changed. The Gatekeeping rule confirmed section
  above already records that no governance gate was touched.
- **README.md**: its only `/uf.unleash` reference is the
  tier-routing row for the Quick tier; README documents no plugin
  tools, checkpoint steps, or gate overrides, and nothing
  description-level changed.
- **docs/**: `docs/configuration.md` covers `uf_delegate_command`
  only for delegation timeout budgets (`timeout`/`idle_timeout`,
  the result envelope, the compaction-storm and final-text
  failure classes) — the new `human_decisions` envelope is not a
  configuration knob and changes no timeout behavior;
  `docs/usage.md` references `/uf.unleash` only at the
  tier-routing and quick-reference level; `docs/cli-reference.md`
  "override" matches are Go CLI flags and per-developer override
  files; `docs/architecture.md`, `docs/heroes.md`, and the
  `docs/decisions/` ADRs mention `/uf.unleash` only at the
  routing/capability level. No docs page documents checkpoint
  override behavior or the human-decision record.

### Documentation-issue gate — applies; orchestrator files it

The AGENTS.md gate (file a documentation issue against the
current repo for user-facing changes before PR merge; exempt:
internal refactoring, test-only, CI-only, spec artifacts) applies
to this change. It is user-facing for every `/uf.unleash` and
`/uf.run-pipeline-plan` operator — new error codes
(`missing_human_decision`, `undeclared_step`,
`escalated_advisory_unrelayed`,
`human_decision_record_unreadable`), a new tool
(`record_human_decision`) and orchestration pattern, and APPROVE
WITH ADVISORIES spec-review results now holding implementation
until relayed — and it matches none of the exemptions. This
matches the `harden-spec-review-gate` precedent (Docs: #1091).
Per the task brief the issue is not created here; the orchestrator
files it before PR and appends the `(Docs: #NNNN)` reference to
the CHANGELOG entry.

### Files touched by 8.4

`CHANGELOG.md` (one new Unreleased/Fixed entry) and this file
only. No rationale failed, so AGENTS.md, README.md, and docs/
were not modified.

## Task 8.5 — Instruction-lint deltas (DE-004)

Executed 2026-11-16 against the implemented working tree at
`e180ae79` (proposal commit, HEAD) plus its uncommitted
phase 2-7 implementation — the same tree task 8.2 describes.
Convention pack: `meta-prompt.md` (MP-011, DE-004 measurable
deltas). Invoked exclusively through the repo wrapper — never
`node` direct — per the task brief:

- Whole-tree current totals: `make lint-instructions`
  (`.uf/scripts/instruction-lint`, no args; 90 files).
- Per-file rows with size deltas: `make lint-instructions
  ARGS="--base e180ae79 --changed"` (changed-files mode;
  `git diff --name-only <merge-base>` plus `ls-files
  --others`; deltas table `lines / tokens / load / status`).
- Fallback verified by extracting the base with `git show
  e180ae79:<path>` into `/tmp/lint-base-compare/` and
  `cmp`/`wc -l`; and by stashing the working tree (`git stash
  push --keep-index`) to re-run the base line whole-tree lint
  — both confirmed the wrapper saw the uncommitted edits;
  no temp-dir re-lint of both ends was needed.

### Per-file rows — 3 canonical files (lint counts this scope)

Source of rows: the `--base e180ae79 --changed` run's
`deltas` table (lines/tokens/load with `(+N)` deltas) plus
its diagnostic header (errors/warnings per file in that
filtered scope). Canonical-to-base subtraction gives the
before values; `git show` + `wc -l` cross-checked the line
counts.

| File | Errors | Warnings in filtered scope | Lines (current vs base) | Tokens (current vs base) | Load (current vs base) |
|------|--------|----------------------------|-------------------------|-------------------------|------------------------|
| `.opencode/commands/uf.unleash.md` | 0 | 2 (IL-030 ~20793 tokens > 8k file budget; IL-031 ~32281 loaded > 20k budget) | 2092 vs 2074 (+18) | 20793 vs 20569 (+224) | 32281 vs 32057 (+224) |
| `.opencode/commands/uf.run-pipeline-plan.md` | 0 | 0 (no diagnostic in changed scope) | 484 vs 464 (+20) | 4941 vs 4696 (+245) | 10303 vs 10058 (+245) |
| `.opencode/agents/cobalt-crush-dev.md` | 0 | 0 (no diagnostic in changed scope) | 284 vs 270 (+14) | 4297 vs 4115 (+182) | 9659 vs 9477 (+182) |

Changed-scope summary line (wrapper-emitted): `total: +52
lines, +651 tokens across 3 files` — verifies as
18+20+14 = 52 and 224+245+182 = 651.

Rows re-measured 2026-11-16 after the code-review fix
loop: iteration 1 added the IL-050 pin annotations and
suppression mechanism and updated the enforcement-code
enumerations in both commands; iteration 2 (finding 4)
extended the run-pipeline human-control-point inventory.
The task-8.5 originals (+13 lines / +128-129 tokens
each) are superseded; same command, same base.

### Scaffold mirrors — byte-identical, lint-excluded by design

The linter's discovery (`find` over `AGENTS.md`,
`.opencode/commands/*.md`, `.opencode/agents/**/*.md`,
`.opencode/skills/**/SKILL.md`, `.opencode/uf/packs/**/*.md`)
does not scan `internal/scaffold/assets/` (confirmed in
`.opencode/lib/instruction-lint.ts:discoverFiles`); mirrors
are therefore absent from lint counts. All three pairs were
verified byte-identical at record time:

- `cmp .opencode/commands/uf.unleash.md internal/scaffold/assets/opencode/commands/uf.unleash.md` — clean
- `cmp .opencode/commands/uf.run-pipeline-plan.md internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md` — clean
- `cmp .opencode/agents/cobalt-crush-dev.md internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` — clean
- `git diff --stat` lists both canonical and mirror edits for
  each pair (`TestEmbeddedAssets_MatchSource` green per task
  8.3).

Their deltas are therefore identical to the canonical rows
above by construction (+14 to +20 lines / +182 to +245
tokens each; +52 lines / +651 tokens across the 3 mirrors;
+104 lines / +1302 tokens repo-wide including mirrors). No
separate lint row exists to report — recording them as
"mirror of canonical, 0 errors / 0 warnings in lint scope,
same deltas" is accurate and avoids double-counting.

### Whole-tree totals and before→after delta

Current (working tree, 2026-11-16, post-fix-loop):

```
.uf/scripts/instruction-lint
90 files: 0 errors, 16 warnings, 20 info (use --all to show info)
```

Base (`e180ae79` stashed, same `make lint-instructions`):

```
90 files: 0 errors, 16 warnings, 20 info (use --all to show info)
```

Delta: **0 errors, 0 warnings, 0 info** — the whole tree
matches base exactly. At task-8.5 record time the delta was
+1 info (`IL-050` `cobalt-crush-dev.md:50`, the not-consent
paragraph pinned near-verbatim across the three prose files
for SYNTH-002); the code-review iteration-1 fix loop added
`instruction-lint-disable-next-line IL-050` pin annotations
at the three block sites (`cobalt-crush-dev.md:50`,
`uf.unleash.md:2078`, `uf.run-pipeline-plan.md:445`) through
the suppression mechanism in `.opencode/lib/instruction-lint.ts`,
recording the intentional duplication mechanically and
returning the info count to base. The 16 warnings are
pre-existing and unchanged — reconciling with the phase-7 run
noted in the brief ("0 errors, 16 pre-existing warnings
(whole tree)"):

- `uf.address-feedback.md` 2 (IL-030, IL-031)
- `uf.fix-ci.md` 1 (IL-030)
- `uf.init.md` 2 (IL-030, IL-031)
- `uf.review-council.md` 2 (IL-030, IL-031)
- `uf.review-pr.md` 2 (IL-030, IL-031)
- `uf.triage-issue.md` 2 (IL-030, IL-031)
- `uf.unleash.md` 2 (IL-030, IL-031) — the only file among the
  three in-scope that carries warnings, before and after
- `go.md` 1 (IL-021 SC-006), `meta-prompt.md` 1 (IL-021 MP-011),
  `python.md` 1 (IL-021 TC-010)

Filtered changed-scope view (`--base e180ae79 --changed`)
reports `90 files: 0 errors, 2 warnings, 1 info` — only the
2 warnings belonging to `uf.unleash.md` and a pre-existing
`IL-050` info on `uf.run-pipeline-plan.md:131` (family prose
shared near-verbatim with `opsx-propose.md`, `speckit.plan.md`,
`speckit.specify.md`; present at base) — confirming no new
warning or info was introduced by the prose edits; the
whole-tree 16 count is identical before and after.

**Prose edits introduced no new lint errors and no new
warnings** — expected and verified. Post-fix-loop the info
count also matches base (20→20; the task-8.5 +1 IL-050 info
is suppressed by its pin annotations); `make lint-instructions`
exits 0 before and after.

### Reconciliation with proposal estimate

Proposal Impact estimated "approximately 7 lines to each of the
three prose files (roughly 21 added lines total, mirrored
byte-identically)". Actual measured delta is **+14 to +20
lines per file (+52 canonical, +104 with mirrors)** — the
not-consent paragraph plus its recovery path, the
enforcement-code enumerations, and the IL-050 pin annotations
exceed the estimate, but the efficiency evidence requirement
(DE-004) is satisfied by recording the measured deltas, not by
hitting the estimate.

### Files touched by 8.5

This file only. No lint config, quality gate, or prose file
was modified for this recording task.

### IL-030 / IL-031 budget growth on `uf.unleash.md`

The two changed-scope warnings are both on `uf.unleash.md`
and both pre-existing at base: IL-030 (self ~20569 tokens at
`e180ae79` vs the 8000-token file budget) and IL-031 (loaded
~32057 vs the 20000-token budget). This change grows them by
+224 tokens (+1.1% self, one-for-one on loaded) through the
pinned not-consent block, its recovery path, and the
enforcement-code enumerations. The budgets are
governance-controlled lint advisories (AGENTS.md
Gatekeeping): raising them is forbidden, and restructuring
`uf.unleash.md` to fit the 8000-token budget is a separate
change outside this scope. Growth recorded here per DE-004 /
Principle VI and acknowledged as an on-record advisory by
the code-review council (iteration 2, findings 5-6: "not a
defect the change introduced").

### Post-8.5 refresh (code-review fix loop, 2026-11-16)

Rows, totals, and whole-tree counts above were re-measured
after the code-review iteration-1 fix loop (IL-050
suppression mechanism in `.opencode/lib/instruction-lint.ts`
— not a scaffolded asset — plus pin annotations and
enumeration updates in both commands) and the iteration-2
finding-4 inventory fix in `uf.run-pipeline-plan.md`.
Scaffold mirrors re-synced byte-identical after each prose
edit. No lint config, quality gate, budget, or scaffold
asset outside the mirrored prose files was modified.
