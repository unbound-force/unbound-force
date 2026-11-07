## Context

PR #786 closed audit finding F6 (session
`ses_e9fe30bfbffez1lcQrExaqQqL2`) with a prose instruction in
`uf.unleash.md`: mechanical-task delegations MUST name the repo
skill or script. Nothing enforces it. The org constitution
states the posture directly: behavioral guards written in prose
are interim mitigations, not the mechanism of record (Principle
VI). Issue #787 asks for the mechanism.

The enforcement surface is `uf_delegate_command`
(`.opencode/plugins/uf-workflow/index.ts`), the single
delegation tool behind quick-tier whole-change implement,
per-task implement delegation, plan and tasks generation
(command-path), and code-fix iterations; the Swarm parallel
path delegates `[P]` tasks through `replicator_forge_spawn_subtask`
and is covered by the template and Guard layers below. Its
input contract is zod-validated with cross-field rules already
(`one of command_path or prompt is required`), its error
envelopes carry codes and remediation, and its checkpoint data
records the delegation facts -- all precedents this design
extends rather than invents.

Three governance artifacts already describe the rule in prose:
the determinism-first skill (classification rubric), the
determinism convention pack DE-001..DE-007 (review rules), and
the Guard reviewer agent (intent drift). None has an
enforcement hook. This change adds one contract and wires the
three layers to it.

## Goals / Non-Goals

### Goals

- A deterministic-classified delegation without a verified
  skill/script reference or a recorded absence reason fails
  closed at the tool, before any child session spawns.
- Every prompt-only delegation declares its classification --
  the silent-omission loophole (just do not classify) closes.
  All three prompt-only call shapes in `uf.unleash.md`
  (quick-tier whole-change, per-task, code-fix) carry it, and
  the Swarm parallel path's spawn prompts carry the
  classification and skill reference through the template
  requirement.
- The classification, resolved reference, and absence reason
  are durable, machine-parseable records (checkpoint data and
  success envelope) that reviewers audit without prompt
  archaeology.
- The rule is citable (DE-008 [MUST], pack v1.3.0) and checked
  (Guard checklist), and the determinism-first skill points at
  the mechanism.
- Judgment work and command-path delegations pass through
  unchanged -- zero false positives by construction.

### Non-Goals

- No semantic classification inside the tool. Deciding whether
  work is deterministic is judgment; the tool validates the
  declared answer (DE-006: judgment at the boundary).
- No change to command-path delegation requirements. The
  command file is the procedure of record; Steps 2/3
  (plan/tasks) keep their current call shape.
- No AGENTS.md behavioral-rule change. The acceptance criteria
  name the skill and the pack; the existing "Determinism gate"
  rule remains accurate without amendment.
- No enforcement against user-level (`~/.config/opencode`)
  skills. The rule binds delegated mechanical work to REPO
  tooling; user-level references are out of scope.
- No tool-contract change to `replicator_forge_spawn_subtask`
  (Replicator-owned, cross-repo). The Swarm path meets the
  contract through the template requirement and Guard
  detection; tool-level validation on that path is a
  Replicator change, tracked by a Replicator-side issue filed
  in task 5.5 so the deferral has a closure path.
- No new plugin, tool registration, or dependency.

## Decisions

### D1: The caller classifies; the tool enforces

`uf_delegate_command` gains three parameters:
`task_classification` (`deterministic` | `judgment`), `skill`
(min 1, max 256 chars), and `skill_absent_reason` (min 1, max
512 chars) -- empty strings are invalid input; absence is
expressed by omitting the field.

`task_classification` is REQUIRED for prompt-only delegation
(no `command_path`) and optional for command-path delegation.
Deterministic classification requires `skill` or
`skill_absent_reason`; `skill` and `skill_absent_reason` are
mutually exclusive; each is valid only with deterministic
classification -- the rendered directive asserts deterministic
work, so judgment-classified and unclassified command-path
calls MUST NOT supply either field.

This mirrors the PR #786 prose semantics exactly -- name the
skill, or record why none exists -- while making both halves
structured. The classification is the determinism-first skill's
twice-test, performed once by the delegating agent and passed
as a validated parameter.

Alternatives rejected:

- **Heuristic skill detection in prompt text.** The tool cannot
  reliably recognize a skill reference in free text, and a
  name-list heuristic rots as skills are added. False
  negatives and positives both; rejected.
- **Warn-only envelope.** The pipeline-hardening audit's lesson
  is that soft warnings lose to context pressure. Fail-closed
  with a structured escape hatch is the house pattern
  (EXPECTED_ARTIFACT_MISSING, INVALID_AGENT).
- **Requiring classification on command-path calls.** Breaks
  the Steps 2/3 call shape for zero enforcement gain: the
  command file already is the procedure of record.

### D2: Skill resolution is an existence check on injectable deps

A `skill` value without `/` resolves as a skill name:
`.opencode/skills/<name>/SKILL.md`, then
`.opencode/skill/<name>/SKILL.md` (both roots exist in this
repo; a name present in both resolves to the first root). A value containing `/` resolves as a repo-relative path
through the existing traversal guard (`resolveRepoRelative`).
(Superseded by `retroactive-push-skill-contracts`,
openspec/changes/retroactive-push-skill-contracts/: the legacy
`.opencode/skill/` root is retired and migrated by `uf init`;
a `/`-free name that misses the skill directory falls back to
a repo-root file reference, and an existing directory is not a
file reference. That change MUST archive after this change.)
Existence is checked with the tool's already-injected
`fileExists` dependency -- no new dependency surface, no real
filesystem in tests.

Resolution failures return `SKILL_NOT_FOUND` (non-retryable)
naming the searched locations, following the
remediation-carrying error-contract precedent
(pipeline-hardening D5).

Alternatives rejected:

- **Embedding the SKILL.md body into the composed prompt.**
  Duplicates content the child loads through its own skill
  tooling and bloats every delegation. The tool injects a
  directive, not the document.
- **A skill registry or manifest.** An existence check on two
  known roots needs no registry; a manifest would add a drift
  surface (the exact failure mode the reviewer-capabilities
  drift caused).

### D3: Prompt injection is a fixed, tool-rendered section

The composed prompt gains a directive section in a fixed
position -- after command text and feature-dir context,
immediately before the extra prompt -- rendered by the tool:

    Required skill: <name-or-path>

    This task is classified as deterministic work. Load the
    skill (invoke the `skill` tool with name "<name>") or run
    the script at <path>, and execute its full procedure.
    Do not substitute an improvised procedure.

Tool-rendered, not agent-reconstructed (DE-003). The exact
wording is a design constant in the plugin, covered by
golden-shape tests.

### D4: Validation splits between schema and execute()

The zod schema enforces every structural rule, so malformed
input is rejected before `execute()` runs: enum membership
(only `deterministic` and `judgment` are accepted; any other
value fails input validation), length bounds (`skill` 1-256,
`skill_absent_reason` 1-512, empty strings invalid), mutual
exclusivity (`skill` and `skill_absent_reason` cannot be
supplied together), each valid only with `deterministic`
classification (judgment-classified and unclassified
command-path calls supplying either field fail validation),
and the prompt-only rule (a
call without `command_path` must declare
`task_classification`). The two rules that need
child-session context or dedicated failure codes --
deterministic classification requires a reference, and the
reference must exist on disk -- execute inside `execute()` and
return the dedicated failure envelopes
(`SKILL_REFERENCE_REQUIRED`, `SKILL_NOT_FOUND`), each
non-retryable with remediation text. Schema-level rules return
generic input-validation failures; execute()-level rules
return the greppable, machine-actionable envelopes (Principle
III), matching how `COMMAND_FILE_READ_FAILED` and
`EXPECTED_ARTIFACT_MISSING` work. All validation happens
before host-model resolution and child-session spawn: a
rejected call costs one tool round-trip, zero sessions.

### D5: Audit trail in checkpoint data and success envelope

The auto-checkpoint's `checkpointData` and the
`DelegateCommandData` success envelope gain
`task_classification`, `skill` (the resolved reference), and
`skill_absent_reason` when present. Both are shallow-merge /
additive changes -- backward compatible for
`uf_unleash_detect_resume` and every existing consumer. This is
the record the Guard audits; it is also what makes a
misclassified delegation visible after the fact.

### D6: Governance layers cite the contract, not restate it

- **Pack**: DE-008 [MUST] in `.opencode/uf/packs/determinism.md`
  (v1.2.0 → 1.3.0), severity mapping updated to include DE-008
  in the MUST/CRITICAL list. Adding a rule strengthens a
  governance gate; no existing rule, severity, or threshold is
  weakened or downgraded (gatekeeping integrity). The
  originating issue (#787) is the documented authorization.
- **Guard**: a Deterministic Delegation Discipline section in
  the Code Review Mode checklist of
  `.opencode/agents/divisor-guard.md` -- self-describing, naming
  the parameter contract rather than citing DE-008 (the agent
  definition is scaffold-embedded; the pack is repo-local),
  covering Swarm-path spawn prompts alongside
  `uf_delegate_command` call sites, and flagging vacuous
  absence reasons (a recorded reason that does not state what
  was searched for or what was found) alongside missing
  references.
  `build_review_prompts` reads agent definitions from disk, so
  the check reaches review prompts with zero plugin changes
  (the determinism-codification precedent).
- **Skill**: the determinism-first body gains the delegation
  rule; Related Enforcement lists DE-008 and the parameter
  contract, closing the loop between rubric and mechanism.
- **Template**: `uf.unleash.md` replaces the "Mechanical tasks
  use repo tooling" prose with a self-describing classification
  step (per task, before delegation -- judgment, legitimately
  prose) and the parameter contract in the call blocks; the
  quick-tier whole-change and code-fix delegations declare
  classifications derived from the nature of the work (the
  twice-test: `deterministic` with `skill` or
  `skill_absent_reason` only when a repo skill or script
  performs the work end-to-end; `judgment` otherwise -- docs
  authorship and prose-finding application are judgment work);
  and the Swarm parallel-path spawn prompts carry the
  classification and skill reference in the subtask
  description.

Alternative rejected: encoding the rule only in the Guard and
pack (layers 1+3 of the issue's options) without the tool
contract. Review-time detection is after the cost is spent;
the tool contract prevents the wasted child session.

### D7: Scaffold mirrors for the three embedded sources

`internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`,
`.../commands/uf.unleash.md`, and `.../agents/divisor-guard.md`
are embedded mirrors; each changes here and each mirror is
updated in the same increment with drift-detection tests green.
The pack and the determinism-first skill are repo-local
(`knownNonEmbeddedFiles`) and ship as source only.
Consequently the three scaffolded sources carry self-describing
guidance: the classification step, the failure-envelope
remediation text, and the Guard checklist name the parameter
contract without citing the repo-local pack or skill, so no
reference dangles in hero repos after `uf init`.

## Risks / Trade-offs

- **Misclassification evades the tool.** A caller can label
  judgment work to skip the requirement. The classification is
  now an explicit, recorded claim the Guard audits against the
  task text; today's alternative is no claim at all. Layered
  mitigation, accepted.
- **Escape-hatch abuse.** `skill_absent_reason` could become a
  free pass. It is recorded verbatim in durable state and is
  the first thing the DE-008 review check reads; vacuous
  reasons (spec-defined: a reason not stating what was
  searched for or found) are flaggable findings. Accepted.
- **Breaking change for prompt-only callers.** Only
  `uf.unleash.md` performs prompt-only delegation today
  (quick-tier whole-change implement, per-task implement,
  code-fix); all three call shapes are updated in this change,
  and the Swarm parallel path's spawn prompts carry the contract
  through the template requirement. Future delegating commands
  must classify -- that is the requirement working as intended.
- **Skill-root hardcoding.** The two roots are this repo's own
  convention; the resolution order is documented and tested.
  A future layout change is a one-constant edit.
- **Downstream lag.** Hero repos receive the tool and template
  via `uf init` sync; unsynced repos keep the old contract and
  keep functioning (Composability First).

Constitution alignment is assessed per principle in
proposal.md. This design is Principle VI applied to its own
enforcement gap, with Principle III shaping the failure
envelopes and audit records.

## Seam verification (DE-007)

This increment binds a prose procedure to a tool contract, so
its tasks.md carries the three-layer seam-verification task:
registration surface (tool args-schema parity with the
documented contract), documented output shape (envelope and
checkpoint fields), and write-then-read round-trips
(checkpoint data readable through `get_progress`).
