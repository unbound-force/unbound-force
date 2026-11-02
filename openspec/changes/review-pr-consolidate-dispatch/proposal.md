## Why

`/uf.review-pr` still mandates manual `finalize_review_dispatch`
payload assembly. The `dispatch-finalize-determinism` change
(0593caeb) moved `/uf.review-council`, `/uf.triage-issue`, and
`/uf.address-feedback` to a single `consolidate_dispatch` call
with provenance; it did not move `uf.review-pr.md`.

Observed in a production `/uf.review-pr` run on dewey PR #138
(session `ses_e66b2decfffe6iTzZlb9ktkM5E`, dispatch
`0274aa47-53a2-4715-9dd0-4951382a6f9a`): 5 runs succeeded, 4
were skipped, 15 findings were submitted. The parent rebuilt the
~50 KB payload by hand three times before the dispatch finalized:

1. Rejected: `payload.runs.N: Unrecognized key: "read_only"` —
   the agent copied a plan-entry field into run records while
   reconstructing the payload.
2. Rejected: `consolidated finding has incorrect contributing
   run ids` / `run finding is missing from consolidated
   findings` — following Section 5's deduplication and
   cross-persona grouping prose, the agent merged the architect
   finding (vault.go:627) and the guard finding (vault.go:625)
   into one consolidated finding. The finalizer requires exact
   correspondence between run findings and consolidated
   findings, so the merged payload could never validate.
3. Accepted only after passing the plugin's 15 consolidated
   findings through unchanged.

This is the same failure class `dispatch-finalize-determinism`
documented (session `ses_eca4843f9ffeKwVLQhSFCrh2Rl`: `read_only`
leaking into run records during manual payload reconstruction).
Root cause in `.opencode/commands/uf.review-pr.md` (mirror:
`internal/scaffold/assets/opencode/commands/uf.review-pr.md`):
Section 5 instructs the parent to deduplicate and apply
cross-persona grouping and compound severity — conflicting with
the finalizer's run-finding correspondence check; Section 5a
uses `consolidate_dispatch` without provenance only to fetch
data; Section 5b mandates the hand-built
`finalize_review_dispatch` call; and no bounded retry policy
exists for mechanical rejections.

A second, smaller defect compounds the risk:
`uf.review-council.md` Section 3 (~L360) still carries "First
deduplicate ... cross-persona grouping" prose left over from the
pre-migration wording. Its Section 5 finalization is already
consolidate-with-provenance, so the prose is harmless to
finalization today — but an agent following Section 3 literally
could reintroduce the same class of finalizer rejection.

## What Changes

1. **One-call finalization for `/uf.review-pr`**: replace
   Sections 5/5a/5b of `uf.review-pr.md` with exactly one
   `consolidate_dispatch({correlation_id, provenance: {branch,
   commit, workflow_id: "uf.review-pr"}})` per finalization
   attempt. Provenance resolves only from the frozen PR input
   context `head_ref`/`head_sha` (as `/uf.address-feedback`
   does), because the finalizer validates the provenance commit
   against `input_context.head_sha`. The command states
   plainly: never assemble, restate, or recompute the payload;
   never call `finalize_review_dispatch` directly. The
   Section 4c dispatch-capture prose that cross-references the
   replaced sections is updated to match.
2. **Bounded retry**: at most two retries, only on mechanical
   rejections, never changing assessment data on retry (the
   existing dispatch-finalization FR-004 policy, applied
   unchanged).
3. **Presentation-only grouping**: cross-persona grouping and
   compound severity remain for presenting the review comment
   only; they MUST NOT touch finalization data.
4. **Single-source proposals and warnings**: the parent reads
   lesson proposals and warnings from the same
   `consolidate_dispatch` result.
5. **Scaffold mirror parity**: the
   `internal/scaffold/assets/opencode/commands/uf.review-pr.md`
   mirror updates byte-identically.
6. **Drift/content tests**: content-gate tests in
   `internal/scaffold/scaffold_test.go` assert the new contract
   and the absence of manual-finalize prose; the
   `uf.review-pr.md` clause set in
   `internal/scaffold/command_contract_test.go` is re-pinned to
   the new one-call contract, and the
   `.opencode/test/golden-output-shape.test.ts` tool-call-site
   inventory is refreshed.
7. **CHANGELOG entry** for the user-facing command behavior
   change.
8. **Review-council Section 3 tidy**: mark the stale
   dedup/grouping prose explicitly presentation-only and sync
   that command's scaffold mirror.

Out of scope: #756 (`speckit.testreview.md` Sections 3.5/3.7
carry the same manual-finalize pattern; separate change) and
#901 (host-source model provenance never resolved in plugin
code; separate cause, not fixed by command edits).

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `dispatch-finalization`: the one-call finalization contract
  extends to the fourth command (`/uf.review-pr`), with
  frozen-PR-context provenance, presentation-only grouping,
  single-source proposals and warnings, mirror parity, and
  content-gate tests; the review-council Section 3 prose is
  tidied to the same presentation-only rule.

### Removed Capabilities
- None.

## Impact

- **Commands**: `.opencode/commands/uf.review-pr.md`
  (Sections 5/5a/5b replaced by one finalization section) and
  `.opencode/commands/uf.review-council.md` (Section 3 tidy);
  both scaffold mirrors under
  `internal/scaffold/assets/opencode/commands/` update
  byte-identically (four files total).
- **Tests**: `internal/scaffold/scaffold_test.go` — new
  content-gate assertions (new contract present, manual-finalize
  prose absent); `internal/scaffold/command_contract_test.go` —
  re-pinned `uf.review-pr.md` correlation-id forwarding clauses;
  `.opencode/test/golden-output-shape.test.ts` — refreshed
  tool-call-site inventory; verified under `make check`.
- **Docs**: CHANGELOG entry; repo-local documentation issue
  (AGENTS.md gate) and `unbound-force/website` tracking issue
  (constitution Cross-Repo Documentation rule), both filed
  before PR merge.
- **Behavior**: no change to verdict semantics, review-comment
  presentation grouping, human-confirmation gates, GitHub
  mutations, or the persisted artifact schema; only the
  finalization path becomes deterministic. No quality or
  governance gate is touched.
- **Resolved issue**: #906 (`originating_issue: 906` recorded in
  `.openspec.yaml` so `/uf.finale` threads `Closes #906`).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Finalization becomes a single tool-authoritative call producing
the self-describing `review-dispatch` artifact and canonical
`review-verdict` v2 projection. The command stops asking agents
to hand-transcribe ~50 KB payloads between tools — the exact
agent-to-agent hand-off that failed three times in the incident
session.

### II. Composability First

**Assessment**: PASS

No new tool, dependency, or service is introduced; the change
reuses the existing `consolidate_dispatch` dual-mode surface
already exercised by three sibling commands. Scaffold assets
stay byte-identical to their `.opencode/` counterparts,
preserving the `uf init` composition path.

### III. Observable Quality

**Assessment**: PASS

Machine-parseable outputs are unchanged in schema and become
reproducible in practice: one deterministic finalize call instead
of three ad-hoc LLM reconstructions. Provenance (branch, commit,
workflow id, correlation UUID, per-run model data) is preserved
by the tool rather than retyped by an agent.

### IV. Testability

**Assessment**: PASS

The new contract is pinned by content-gate tests asserting both
the presence of the one-call pattern and the absence of
manual-finalize prose, alongside the existing byte-identical
mirror drift tests; all run under `make check` in isolation
with no external services.

### V. Security by Default

**Assessment**: PASS

The change strengthens the fail-closed boundary: retry
classification stays deterministic and default-deny (the
existing FR-004 policy), provenance resolves only from validated
frozen input context, and indeterminate failures fail closed to
`INCONCLUSIVE`. No new input surfaces, secrets, file
permissions, or privilege boundaries are introduced.

### VI. Deterministic Execution

**Assessment**: PASS

The finalization procedure — previously judgment-reconstructed
per attempt — becomes a deterministic single tool invocation
with a bounded, mechanically classified retry policy, removing
the improvised payload assembly that produced divergent
outcomes across identical retries.
