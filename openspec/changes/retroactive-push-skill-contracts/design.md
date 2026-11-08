# Design: retroactive-push-skill-contracts

## Context

PR 948 (`quick/uf-workflow-tool-fixes`) shipped three contract
changes under the quick tier. Round-2 review found (a) the
changeset tripped MP-009 risk floors (data-loss path in
`migrateSkillDir`; plugin-tool input contract broken by
required `branch`) with the design disposition recorded only in
code comments and CHANGELOG prose, and (b) five live unarchived
OpenSpec deltas mandate the superseded behavior:

- `delegation-skill-enforcement/specs/unleash/spec.md` (L31-41)
  and `model-matrix-delegation/specs/unleash/spec.md` (L52-62,
  plus the L181-188 collision scenario): two-root skill
  resolution (`.opencode/skills/` then `.opencode/skill/`).
- `delegation-skill-enforcement/design.md` (L120) and
  `tasks.md` (L42): prose carrying the legacy root.
- `finale-token-reduction/specs/plugin-tools.md` (L72-96):
  optional-`branch` `git_push` and no-argument
  `git_push_force_with_lease` returning `{success: true}`.
- `port-uf-rebase-pr/specs/uf-rebase-pr/spec.md` (L255-296):
  raw shell `git push --force-with-lease <HEAD_REMOTE>
  <HEAD_BRANCH>` with a plugin-tool SHOULD and raw fallback.
- `spawn-subtask-classification-params/specs/determinism-policy/
  spec.md` (L188-196): an illustrative absence reason naming
  the retired `.opencode/skill/` root.

The live specs tree (`openspec/specs/`) carries no
unleash/plugin-tools/uf-rebase-pr capabilities yet — all five
deltas are unarchived, so requirements apply archive-order-
dependent against sibling deltas' text. The precedent for
superseding text inside the unarchived stack is
`accept-variantless-parent-model` (MODIFIED restatement +
inline annotations + archive-order constraint).

## Goals

- Record the shipped design disposition (MP-008 duty) for the
  push-tool, skill-resolution, and migration contracts.
- Bring live delta text back in line with shipped behavior so
  archive-time materialization cannot resurrect superseded
  contracts.
- Make the supersession discoverable from the contradicted
  deltas themselves (annotations travel with the text).

## Non-Goals

- No code changes of any kind (code shipped and council-
  reviewed in PR 948).
- No new FR ids (existing allocations carry forward).
- No restructuring of the older deltas beyond light inline
  annotations.
- No change to quality gates, CI, review limits, or severity
  definitions.

## Decisions

### D1: Single skill root plus repo-root file fallback

**Decision**: Skill resolution probes
`.opencode/skills/<name>/SKILL.md`, then falls back to a
repo-root file named `<name>` (a directory is NOT a file
reference). Values containing `/` remain traversal-protected
repo-relative paths. Whitespace-bearing values fail
`SKILL_NOT_FOUND` with commands/make-targets guidance.

**Rationale**: The legacy `.opencode/skill/` root is retired by
the `uf init` migration (#933); keeping it in the resolver
would resurrect the two-root ambiguity the migration removes.
The repo-root fallback (#934) covers deterministic references
to root-level scripts (`Makefile`, `build-all.sh`) without
widening the surface to directories. Whitespace rejection is
the accepted tradeoff: a spaced value is far more likely a
mistaken command or make target than a path, and the failure
message says so.

**Constitution alignment**: Deterministic Execution — the
resolver's behavior is fully pinned and test-covered; Spec-
Driven Development — the tradeoff is now recorded where
reviewers expected it.

### D2: Required branch, nickname remote, structured results

**Decision**: `git_push` and `git_push_force_with_lease`
require `branch`, which MUST be the currently checked-out
branch (`BRANCH_MISMATCH` refusal otherwise, including detached
HEAD and indeterminable-current-branch cases). `remote` is
optional, defaults to `origin`, and rejects `/` (git would
resolve a slash-bearing push target as a local repository
path). Results are structured `GitPushData` (`{remote, ref,
old_sha, new_sha, updated, warning?}`).

**Rationale**: The no-argument shape pushed whatever the
checkout's default refspec chose — after a checkout fallback
that can be the wrong branch. The explicit refspec pins the
push target; the mismatch guard refuses rather than guessing;
nickname validation closes the local-path misresolution. The
structured result replaces `{success: true}` so callers can
chain (e.g. `expected_sha` sequencing from the returned
`new_sha`).

**Constitution alignment**: Deterministic Execution — push
targets are explicit and guarded, not inferred.

### D3: Warn-not-fail post-push and removal semantics

**Decision**: When a push lands but `setUpstream` tracking
configuration fails, the tool returns success with a `warning`
carrying the remediation command. When `uf init`'s
skill-directory migration cannot remove a legacy file, the
migration warns and keeps the file rather than counting it
removed.

**Rationale**: The push (the operation the caller asked for)
succeeded; failing the whole call for a tracking-config
follow-up would misrepresent the state and invite a redundant
force-push retry. Symmetrically, a failed removal must not be
reported as a successful migration: the warning makes the
residue visible without aborting the whole migration over one
file.

**Constitution alignment**: Observable Quality (III) — the
structured `warning` field surfaces a partial failure in
machine-parseable output with its remediation instead of
silently misreporting success.

### D4: Always-plugin-tool force-push in rebase-pr

**Decision**: `/uf.rebase-pr` always calls
`git_push_force_with_lease` with `{branch: <HEAD_BRANCH>,
remote: <HEAD_REMOTE>, expected_sha: <HEAD_SHA>}` — including
after the fallback checkout — and raw shell force-push forms
are forbidden in the command.

**Rationale**: The old SHOULD (plugin tool when upstream
matches, raw shell otherwise) forked the push path on upstream
configuration — exactly the state a fallback checkout
invalidates. The tool's explicit refspec makes the push
independent of upstream configuration, and its lease on the
PR-reported head SHA bounds the push. `expected_sha` sequencing
(on first force-push pass the PR-reported head, then the
previous push's returned `new_sha`) removes reliance on the
remote-tracking ref, which fork PRs lack.

**Constitution alignment**: Deterministic Execution — one push
path, bounded by the tool's guards, test-pinned by
`TestCommandContracts_PushToolInvocationPins`.

### D5: Retroactive record strategy

**Decision**: One retroactive change with (1) MODIFIED
restatements of the three contracts from the latest-generation
delta text, (2) light inline superseded-by annotations on each
of the five contradicted deltas (fix-tui-git-output-leak
precedent), and (3) archive-order constraints stated in four
discoverable places: the delta header comments, the requirement
prose context, proposal.md (What Changes + Impact), and this
design.

**Rationale**: Restatement is the only OpenSpec-supplied
supersession mechanism (requirement-name matching; MODIFIED
replaces wholesale). Annotations make the supersession visible
from the contradicted text itself, surviving archive. OpenSpec
has no mechanical cross-change ordering guard, so the
constraint must be recorded redundantly in every place a future
archiver will look.

**Constitution alignment**: Autonomous Collaboration (I) — the
retroactive record is the collaboration artifact: restatements,
superseded-by annotations, and archive-order constraints are
recorded where a future archiver will look, discoverable from
the artifacts themselves without runtime coordination. The
strategy is the accepted remedy from the round-2 findings.

## Risks / Trade-offs

### Risk: Archive ordering has no mechanical guard

A future archiver could archive this change before an
annotated sibling, letting the sibling's superseded text
materialize into the live spec.

**Mitigation**: The constraint is stated in the delta header
comments, the requirement prose, proposal.md, and design.md —
the four surfaces an archiver consults. The scenario-
continuity abort provides a backstop at this change's own
archive.

### Risk: Restatement drift until archive

Until this change archives, the live deltas still carry the
contradicted text; a reader of `model-matrix-delegation` alone
sees two-root semantics.

**Mitigation**: The inline annotations mark every contradicted
span with a pointer here, so the drift is labeled rather than
latent.

### Risk: Pre-existing stack question

"Command delegation tool" is only ever MODIFIED in the
unarchived stack (delegation-skill-enforcement, then
model-matrix-delegation); which change originally ADDED the
requirement is not materializable while those remain
unarchived.

**Mitigation**: Documented here as out of scope; the numbering
header in model-matrix-delegation already records the
retroactive UF-FR-002 allocation for the same reason.

### Risk: finale-token-reduction delta is mis-pathed

Discovered while validating this change: finale-token-reduction
carries its plugin-tools delta at `specs/plugin-tools.md` — a
single file directly under `specs/` — instead of
`specs/plugin-tools/spec.md`. `openspec validate --strict`
reports this as a pre-existing ERROR ("Delta specs must be a
spec.md inside a capability folder — this file is ignored when
the change is applied or archived"), meaning that change's
requirements (including the original `Git operation tools`
ADDED) never materialize into the live spec as-is.

Consequence for this change: the plugin-tools MODIFIED here
can only archive once `Git operation tools` exists in the live
spec, which additionally requires finale-token-reduction's
delta to be re-pathed to `specs/plugin-tools/spec.md` before
(or at) its archive. Repairing another change's file layout is
outside this change's scope (annotations only, per the
reviewer remedy).

**Mitigation**: Documented here and in the plugin-tools delta
header's archive-ordering note; the re-path is a mechanical
file move the archiver performs when finale-token-reduction
archives. Without it, the archive of this change's
plugin-tools delta is refused (target spec does not exist) and
fails safely rather than silently.

### Trade-off: Whitespace restriction in skill references

Spaced file paths are not valid skill references — a value
like `make plugin-test` fails `SKILL_NOT_FOUND` with guidance
instead of resolving. Recorded as accepted in D1 and pinned by
the UF-FR-002 restatement; the alternative (path-parsing
spaced values) would reintroduce command/make-target ambiguity
the deterministic-delegation rule exists to remove.
