## Why

Issue #735 documents that `.opencode/commands/uf.triage-issue.md`
specifies dispatch data flows the review-dispatch tools do not return
to the calling command:

1. **Section 2.3** claims `dispatch_review_runs` returns `status`, run
   counters, consolidated `findings`, `proposals`, and `warnings`. The
   caller-visible output is a one-line counts summary
   (`dispatch_review_runs: ok — 7 completed, 0 failed, 3 skipped, 6
   findings, 6 proposals`); the arrays live only in tool metadata the
   calling model never sees.
2. **Section 2.4** consumes a `proposals` array that never reaches the
   caller — and instructs reading a `reference` field that does not
   exist (the validated proposal shape is `{agent, information, tag,
   category}`, verified at `review-dispatch/index.ts:2973` and
   `:3239`).
3. **Phase 3 (Section 3.1)** requires consolidating per-persona
   "successful model-run verdicts" plus category, objectivity,
   reasoning, and split recommendations, but names no data source for
   either. Run-recorded `workflow_verdict` values are surfaced only by
   `consolidate_dispatch`; the assessment contract fields exist only
   in child session transcripts.
4. **Section 4.6** frames `consolidate_dispatch` as finalization-only
   ("exactly once per finalization attempt, with provenance as its
   only argument"), leaving its retrieval mode (called without
   provenance it returns the raw payload) invisible to the command.

Observed consequences (triage of fullsend-ai/fullsend#8465: 7 runs, 3
content-persona skips): the executing agent improvised retrieval via
`list_sessions` + `db_read_session` across all 7 child sessions
because no sanctioned path was documented, and encountered two
disagreeing verdict sources — all 7 child text assessments said VALID
while run-recorded kind-native `workflow_verdict` said 5
NEEDS-CLARIFICATION / 2 VALID — then had to infer from Section 4.6's
finalizer-authoritative clause that the run-recorded source governs.
Improvised retrieval and inferred precedence are correctness and
determinism risks: a different execution could resolve the same
dispatch to a different verdict.

The identical false claims from items 1 and 2 also appear in the
sibling commands `uf.review-council.md` (Section 4),
`uf.address-feedback.md` (Section 2.3.7), and `uf.review-pr.md`
(Section 5a): all four commands tell the caller that
`dispatch_review_runs` returns a `proposals` array and that proposals
carry a `reference` field. Both claims are false everywhere they
appear.

## What Changes

Docs-first correction (Option B in the issue), expanded after spec
review remediation: make the commands document the data flows the
tools actually support, plus one determinism-gate extraction — the new
`read_assessment_fields` plugin tool that owns the assessment-field
reads and their untrusted-input validation as executable, tested code
(design D4). Two maintainer-approved scope additions ride in this
change: the same proposal-sourcing correction applied to the three
sibling commands, and the adopted `@opencode-ai/plugin` 1.4.10 →
1.19.0 pin update with its lockstep set (see "Dependency Pin Update"
below).

Edits to `.opencode/commands/uf.triage-issue.md` (canonical) mirrored
byte-identically to `internal/scaffold/assets/opencode/commands/
uf.triage-issue.md`:

1. **Section 2.3** — replace the "tool returns ... consolidated
   findings, proposals ... and any warnings" claim with the actual
   contract: a counts-only summary line. Add a sanctioned retrieval
   step: immediately after `dispatch_review_runs` completes, call
   `consolidate_dispatch` once WITHOUT provenance to fetch the raw
   payload (per-run records with run-recorded `workflow_verdict`,
   findings, advisories, run counts, and the top-level `proposals`
   array). Document the failure behavior: a transient retrieval
   failure retries the side-effect-free call exactly once then fails
   closed; a permanent failure fails closed without retry; an all-skip
   dispatch (zero completed, zero failed runs, no dispatch session) is
   recorded as a valid outcome per the no-vote semantics, not treated
   as a retrieval failure.
2. **Section 2.4** — source the proposals array from the
   no-provenance `consolidate_dispatch` payload instead of
   `dispatch_review_runs`. Per-proposal `dewey_store_learning`
   handling is unchanged except for a field-name correction: the
   validated proposal shape is `{agent, information, tag, category}`
   (there is no `reference` field), so the learning category comes
   from the proposal's `category` field.
3. **Section 3.1 (Phase 3)** — name the authoritative data sources:
   - Verdict resolution uses run-recorded `workflow_verdict` values
     from the consolidate payload's `runs` (the tool-computed,
     kind-native source), not child prose.
   - Category, objectivity, reasoning, and split recommendations come
     from a single `read_assessment_fields` call with the dispatch's
     run identifiers (item 4 below); the command performs no
     transcript reads and no per-value validation of its own.
   - When child transcript text and run-recorded `workflow_verdict`
     disagree, the run-recorded verdict governs.
   - When a run's persisted child text is missing or unparseable, the
     tool reports that persona's assessment fields as unavailable;
     verdict resolution proceeds from the run-recorded values.
4. **Assessment-field extraction tool** — implement
   `read_assessment_fields` in the review-dispatch plugin
   (`.opencode/plugins/review-dispatch/index.ts`) with unit and
   integration tests (tasks 2.1-2.4): given the dispatch's run
   identifiers, the tool reads each run's persisted child final text
   from the dispatch store, parses the structured response contract
   fields (category, objectivity, reasoning, split recommendations),
   enforces the untrusted-input validation (Constitution V) — length
   bounds (`category` non-empty and at most 128 characters;
   `objectivity` at most 64; `reasoning` at most 4096; split
   recommendations at most 16 entries of at most 512 characters
   each), type checks, and contract-shape validation, skipping values
   that fail as unavailable — and returns per-run validated fields.
   This extracts the mechanical procedure the determinism pack's
   DE-001 forbids as command prose; no session-database reads
   (`list_sessions` / `db_read_session`) remain in the documented
   flow.
5. **Section 4.6** — document `consolidate_dispatch`'s dual mode:
   without provenance it is the sanctioned retrieval path (no
   finalization side effects); with provenance it finalizes. The
   "exactly once per finalization attempt" bound applies only to the
   provenance-bearing finalization call; the earlier retrieval call
   from Section 2.3 does not count against it.
6. **Sibling commands** — correct the shared Store Lesson Proposals
   contract per command, matching each command's finalization
   architecture (FR-005): `uf.review-council.md` (Section 4) and
   `uf.address-feedback.md` (Section 2.3.7) source proposals from
   the no-provenance `consolidate_dispatch` retrieval payload, with
   a dual-mode note added to their finalization sections
   (review-council Section 5; address-feedback Section 2.3.8)
   scoping exactly-once to the provenance-bearing call;
   `uf.review-pr.md` (Section 5a) adds the same no-provenance
   retrieval step after its `dispatch_review_runs` call (the
   command dispatches once, so a dispatch session exists and the
   retrieval is available) while its Section 5b direct
   `finalize_review_dispatch` finalization is unchanged. In all
   three, the learning category is the proposal's `category` field,
   not a nonexistent `reference`.
7. **Dependency pin lockstep** — the maintainer-approved
   `@opencode-ai/plugin` 1.4.10 → 1.19.0 bump already present in
   the working tree is adopted into this change with the full
   lockstep set (scaffold asset mirrors, pin-enforcement test
   expectation, provenance testdata, AGENTS.md
   Technology Stack and README, upstream verification evidence) and
   validation evidence; see "Dependency Pin Update".

No existing plugin tool code, schemas, or outputs change; the only
plugin-side edit is the additive `read_assessment_fields` tool (item
4). No CI workflow, coverage, or severity-threshold changes; the only
gate-adjacent edits are the pin-enforcement test expectation that
tracks the approved dependency pin and the new plugin tool's test
suite additions.

## Capabilities

### Modified

- `triage-issue`: the command's dispatch data-flow contract — where
  proposals, run verdicts, and assessment fields come from (the
  `consolidate_dispatch` retrieval path and the new
  `read_assessment_fields` extraction tool), and the dual-mode use of
  `consolidate_dispatch`.
- `review-council`, `address-feedback`, `review-pr` (shared Store
  Lesson Proposals contract): the same proposal-sourcing and
  field-name corrections, tracked in this change's tasks rather than
  separate capability deltas.

### New

None.

### Removed

None.

## Impact

- **Files**: `.opencode/commands/uf.triage-issue.md` and its
  byte-identical scaffold mirror
  `internal/scaffold/assets/opencode/commands/uf.triage-issue.md`;
  the three sibling command files and their scaffold mirrors;
  `.opencode/plugins/review-dispatch/index.ts` (new
  `read_assessment_fields` tool) and the plugin test suite under
  `.opencode/test/`; `.opencode/package.json`
  (approved pin, `engines`
  declaration) and its scaffold asset
  mirror; the
  `TestPluginPackageInputs_ExactVersions`
  expectedRuntime map in `internal/scaffold/scaffold_test.go`;
  `internal/scaffold/testdata/plugin-package-provenance.json`
  (pin-governance testdata consumed
  by `TestPluginPackageInputs_ApprovedMaintenanceProvenance`); `AGENTS.md`
  Technology Stack and `README.md` (Node engine floor); the
  `review-council-multi-model-fanout` delta (superseded-by note);
  `CHANGELOG.md` (drift-detection tests in
  `internal/scaffold/scaffold_test.go` enforce command asset parity
  under `make check`).
- **Behavior**: the command-doc edits align documented data flows
  with existing tool behavior; Phase 3's assessment-field sourcing
  changes from prose-prescribed transcript validation to a call to
  the new tested `read_assessment_fields` tool (verdict resolution
  is unchanged — it already resolves from run-recorded values); the
  dependency pin update changes the plugin SDK runtime from 1.4.10
  to 1.19.0 (validated by `make plugin-test`).
- **Sibling commands**: review-council, address-feedback, and
  review-pr carry the identical false claims and are IN scope
  (tasks 1.5-1.7). `speckit-testreview` does not consume the
  proposals array (verified: no `proposals` references in its
  command file) and stays out of scope.
- **Risks**: moderate. The command-doc edits remain low-risk and
  gated by scaffold asset parity. The new plugin tool is additive
  (no existing tool contract changes) and gated by its unit and
  integration tests under `make plugin-test`. The dependency pin
  update touches a governance-gated pin (see "Dependency Pin
  Update"), widens the review surface, and ships its lockstep file
  group as a dedicated commit (task 4.5) per the AGENTS.md
  commit-scope rule.
- **Documentation**: both documentation gates satisfied — repo
  documentation issue unbound-force/unbound-force#754 and org
  Cross-Repo Documentation issue unbound-force/website#312
  (referencing companion tracker website#303 for the antecedent
  change), covering the command-doc corrections and the Node
  engine-floor documentation surface.

## Dependency Pin Update (maintainer-approved)

The working tree carries `@opencode-ai/plugin` 1.4.10 → 1.19.0
(`.opencode/package.json` plus regenerated lock; `node_modules`
already installed at 1.19.0). Spec review flagged it as an
out-of-scope, governance-gated change; the maintainer elected to keep
it in this change rather than revert. Justification and obligations:

- **Why**: 1.4.10's tool-argument handling silently stripped runtime
  arguments — the defect the `fix-tool-args-schema-mismatch` change
  works around locally, with live `CHECKPOINT_ERROR` failures
  documented in its retrospective. 1.19.0 tracks the current OpenCode
  plugin SDK line so the local workaround is no longer the only
  correct path forward.
- **Upstream verification** (task 3.3; AGENTS.md Gatekeeping
  condition): verify the target artifacts against the upstream
  releases — npm registry integrity and each package's actual
  upstream license: MIT for `@opencode-ai/plugin` 1.19.0 and
  `@opencode-ai/sdk` 1.19.0; Apache-2.0 for `@ai-sdk/provider`;
  AFL-2.1 OR BSD-3-Clause for `json-schema` 0.4.0 (do not assume MIT
  across the set) — following the org precedent set by
  `review-council-multi-model-fanout` task 1.5, and record the
  evidence in `internal/scaffold/testdata/`
  `plugin-package-provenance.json` (plugin 1.19.0, fresh
  `verified_at`, registry-maintainer and HTTPS provenance). The
  msgpackr 1 → 2 major transitive bump (resolved 2.1.1) is verified
  before adoption: review the 2.x changelog for breaking changes
  affecting the plugin runtime and confirm `make plugin-test` green
  (task 4.3).
- **Engine floor check** (tasks 4.1 and 4.4; AGENTS.md Gatekeeping
  condition): ini 7 raises the Node engine floor to a disjunctive
  range (^22.22.2 || ^24.15.0 || >=26.0.0). The range is verified
  by a locally executed, reproducible check (assert the manifest
  `engines.node` equals ini 7's declared range and record the
  result), declared in the `engines` field of
  `.opencode/package.json` and its scaffold asset mirror (task 4.4),
  and documented in AGENTS.md and README (task 4.2) — not deferred
  to CI's floating `node-version: 22`.
- **Evidence**: tasks 3.3 and 4.3 record the upstream verification,
  `make plugin-test` green against 1.19.0, and the pin-enforcement
  test green after the lockstep update.
- **Lockstep** (tasks 3.4-3.7, 4.2): scaffold asset mirrors
  (`internal/scaffold/assets/opencode/package.json`), the
  `TestPluginPackageInputs_ExactVersions`
  expectedRuntime map, the pin-governance testdata
  (`plugin-package-provenance.json` at 1.19.0 with fresh
  `verified_at`), and the AGENTS.md Technology Stack and README prose.
- **Rollback**: the 1.4.10 baseline remains recoverable from git
  history (the pin lockstep commits are revertable); revert the
  coupled file groups (manifest, asset mirrors, test expectations,
  testdata, AGENTS.md and README lines) in one commit.
- **Transitive risk notes**: 1.19.0 pulls `@opencode-ai/sdk` 1.19.0
  and updates effect/ini/msgpackr transitively (msgpackr 1 → 2,
  resolved 2.1.1, verified per the upstream-verification bullet).
  The uuid 13.0.1 override (pre-existing `overrides` entry in
  `.opencode/package.json`, unchanged by this bump) remains in
  force even though effect declares ^14.0.0: the override predates
  this change and masks effect's ^14.0.0 range tree-wide for the
  plugin test runtime, forcing uuid 13.0.1 there; its continued
  safety is demonstrated by the recorded validation evidence (task
  4.3) rather than altered here.
- **Supersession note**: SC-FR-003 in the unarchived
  `review-council-multi-model-fanout` change recorded the 1.4.10 pin
  for that change's verification; this change supersedes the
  repo-wide pin per maintainer decision, the live enforcement
  (the Go pin test) is updated in lockstep, and task 5.5 adds the
  superseded-by annotation to the affected delta.

## Constitution Alignment

- **I. Autonomous Collaboration — PASS**: The change replaces
  improvised runtime retrieval (session-database spelunking) with a
  sanctioned, artifact-based data-flow contract, so the command and
  dispatch tools interact through defined interfaces rather than
  ad-hoc coupling.
- **II. Composability First — PASS**: The change adds one plugin
  tool (`read_assessment_fields`) without modifying any existing
  tool contract, so every hero remains independently installable
  and usable. The fix documents the existing interface and extends
  it additively.
- **III. Observable Quality — PASS**: Naming authoritative sources
  makes verdict resolution traceable to run-recorded, tool-computed
  values with provenance instead of ambiguous prose, and removes an
  inferred precedence rule.
- **IV. Testability — PASS**: The command-doc edits are verified by
  the existing scaffold drift-detection tests (`make check`); the
  new extraction tool is verified by plugin unit and integration
  tests (`make plugin-test`); the dependency pin update is verified
  by `make plugin-test` and the pin-enforcement test.
- **V. Security by Default — PASS**: The pin update keeps exact
  version pinning with lockfile integrity hashes and per-package
  license verification (MIT for `@opencode-ai/plugin` and
  `@opencode-ai/sdk`; Apache-2.0 for `@ai-sdk/provider`; AFL-2.1
  OR BSD-3-Clause for `json-schema`); transitive risk notes
  (msgpackr major bump, uuid override) are recorded above and
  validated by the plugin test suite and CI.
- **VI. Deterministic Execution — PASS**: The change replaces
  improvised retrieval with a documented, deterministic data flow
  (one retrieval call, run-recorded verdict precedence) and
  extracts the assessment-field reads and untrusted-input
  validation into the tested `read_assessment_fields` tool (design
  D4), so no mechanical validation procedure remains as command
  prose. The determinism pack's DE-001 is satisfied by extraction
  rather than bounded prose; design decisions D4 and D7 record the
  engagement and its resolution.
