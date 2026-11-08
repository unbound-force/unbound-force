# Proposal: retroactive-push-skill-contracts

## Why

PR 948 (`quick/uf-workflow-tool-fixes`, review round 2) shipped
three contract changes to the `uf-workflow` plugin and its
commands: the push tools now require `branch` and push via an
explicit refspec with structured results (#947), skill
resolution dropped the legacy `.opencode/skill/` root in favor
of the single `.opencode/skills/` root plus a repo-root file
fallback (#934), and the `uf init` skill-directory migration
treats removal failures as warnings rather than silent
successes (#933). The review council's round-2 findings
identified two documentation obligations this change
discharges:

1. **Tier-routing risk floors (divisor-guard, HIGH).** The
   changeset rode the quick tier while touching a data-loss
   path (`migrateSkillDir`) and breaking the plugin-tool input
   contract (required `branch`), both of which match the risk
   floors that forbid the quick tier (MP-009). The merge/lease
   and skill-resolution semantics shipped as code comments and
   CHANGELOG prose with no reviewed design record. MP-008
   requires hotfixed behavior to be documented retroactively.
2. **Spec consistency (divisor-architect, MEDIUM).** Five live
   unarchived OpenSpec deltas mandate the superseded behavior:
   `delegation-skill-enforcement` and `model-matrix-delegation`
   require two-root skill resolution;
   `finale-token-reduction` pins the old no-argument push
   contracts; `port-uf-rebase-pr` mandates the deleted raw
   shell force-push fallback; `spawn-subtask-classification-
   params` illustrates an absence reason against the retired
   root. Live deltas contradicting shipped contracts fail the
   zero-waste rule at archive time.

**Workflow-record note** (precedent:
`accept-variantless-parent-model`, proposal.md "Why"): this
change rides the quick-tier branch `quick/uf-workflow-tool-fixes`
because the council fix loop on that branch demanded a durable
spec record for already-shipped behavior; re-routing the fix
loop to a fresh OpenSpec branch would orphan the fix-loop
history. The tier divergence is documented here as a
council-demanded escalation, exactly as the precedent recorded.

## What Changes

- Adds MODIFIED requirement deltas restating three contracts
  to the shipped behavior (no new requirement ids; existing
  allocations and name addressing carry forward):
  - `specs/unleash/spec.md` — "Command delegation tool"
    (UF-FR-002): single skill root `.opencode/skills/` with a
    repo-root file fallback (a directory is not a file
    reference), whitespace-bearing values rejected with
    commands/make-targets guidance; the stale two-root
    collision scenario is replaced by root-file fallback
    scenarios; all other scenarios carry forward verbatim.
  - `specs/plugin-tools/spec.md` — "Git operation tools":
    `git_push` and `git_push_force_with_lease` require
    `branch` (the currently checked-out branch, pushed via an
    explicit refspec, `BRANCH_MISMATCH` refusal), `remote` is
    a validated git nickname (rejects `/`), structured
    `GitPushData` results with an optional post-push
    `warning`; all other git tools restate verbatim.
  - `specs/uf-rebase-pr/spec.md` — "Bounded force-push": the
    command always calls `git_push_force_with_lease` with
    `{branch: <HEAD_BRANCH>, remote: <HEAD_REMOTE>,
    expected_sha: <HEAD_SHA>}`; raw shell force-push forms are
    forbidden.
- Adds light inline superseded-by annotations at the five
  contradicted live deltas (fix-tui-git-output-leak precedent)
  so the supersession record travels with the text when those
  changes archive.
- Records archive-order constraints: this change MUST archive
  after `delegation-skill-enforcement`,
  `model-matrix-delegation`, `finale-token-reduction`, and
  `port-uf-rebase-pr` (and, for annotation continuity, after
  `fix-tui-git-output-leak`), stated in the delta headers,
  this proposal, and design.md.
- Records the accepted whitespace-restriction tradeoff: spaced
  file paths are not valid skill references; the failure
  envelope points at commands/make-targets guidance instead.

## Capabilities

### Modified Capabilities

- `unleash` — Command delegation tool skill-resolution
  contract (single root, repo-root file fallback, whitespace
  rejection).
- `plugin-tools` — Git operation tools push contracts
  (required branch, nickname remote, structured results,
  warning field).
- `uf-rebase-pr` — Bounded force-push strategy
  (always-plugin-tool, structured arguments).

### New Capabilities

None.

### Removed Capabilities

None.

## Impact

- **Spec record only.** No code changes: the implementation
  shipped and passed council review in PR 948 (commits through
  the round-2 fixes). This change is the durable design
  disposition MP-008 requires.
- **Live-delta consistency.** The five contradicted deltas
  carry inline annotations; at archive time their requirement
  text flows into the live specs already superseded, and this
  change's restatements replace it. The archive-order
  constraint prevents the reverse ordering (an older delta
  archiving later and overwriting the corrected text).
- **CHANGELOG.** Already amended in PR 948 (push-tool BREAKING
  guidance, branch-probe distinction, remote validation,
  warning field, removal-failure semantics; Docs references).
- **Documentation.** Docs issues filed: unbound-force#963
  (in-repo) and unbound-force/website#379 (website; notes
  stale website#310/#317). No further docs surface changes.
- **Scaffold assets.** None: OpenSpec changes are not
  scaffolded assets.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0).
This is a pure spec-record change: deltas, annotations, and
archive-order constraints bring the spec surface back in line
with shipped behavior. It discharges prior council findings
(the round-2 tier-routing and spec-consistency items) through
their accepted remedies, changes no council-reviewed code, and
passes through the council gate again as part of the PR.

### I. Autonomous Collaboration

**Assessment**: PASS. The change records the design rationale
for contracts the swarm's review council already demanded be
recorded, closing the loop between review findings and the
spec record without re-opening settled code. The
retroactive-documentation duty (MP-008), triggered when review
finds a quick-tier changeset tripped risk floors, is discharged
through self-describing spec artifacts that future heroes
consume asynchronously.

### II. Composability First

**Assessment**: PASS. The restated plugin-tool and
skill-resolution contracts are extension points every hero
repository consumes standalone through its scaffolded command
copies; the change introduces no mandatory cross-hero
dependency and requires no hero to patch or fork another to
integrate.

### III. Observable Quality

**Assessment**: PASS. The restated contracts pin the exact
structured tool arguments, structured outputs ({remote, ref,
old_sha, new_sha, updated}), lease semantics, and failure
codes the implementation enforces, keeping push and
skill-resolution behavior inspectable and test-pinned rather
than convention-only.

### IV. Testability

**Assessment**: PASS. Every restated contract cites its
enforcing tests: TestCommandContracts_PushToolInvocationPins
census-pins the push call sites across uf.finale, uf.fix-ci,
and uf.rebase-pr; the uf-workflow INVALID_INPUT tables cover
the branch and remote schemas; migrate_skill_test.go covers
the skill-dir migration failure paths.

### V. Security by Default

**Assessment**: PASS. The recorded contracts preserve the
fail-closed postures the implementation enforces: the
force-push lease on the pre-push head SHA, the traversal guard
on repo-root skill references, and RemoteNameSchema's
rejection of local-path push targets. No security control is
weakened; the accepted whitespace-restriction and
warn-not-fail tradeoffs are recorded with their rationale.

### VI. Deterministic Execution

**Assessment**: PASS. No gates are modified: coverage
thresholds, CI flags, review limits, and severity definitions
are untouched. The change pins the deterministic resolution
order (skill name first, then repo-root file) and the exact
push-tool arguments, removing convention-only ambiguity from
execution surfaces.
