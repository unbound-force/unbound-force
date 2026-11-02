## Context

Issue #914, the follow-up recorded in `host-fallback-provenance`
design D2 ("Advisor-symmetry conflation", Fixes #901):
`resolveHostModel()` failures cover two causes with opposite
remedies, but the shared dispatch core classifies all of them
identically on both branches that call it.

`StructuralHostModelError` (`.opencode/lib/agent-executor.ts`,
L231) marks causes retrying cannot fix — the parent message is not
an assistant message, carries no model identity, or carries an
invalid variant — and carries a `remediation` string naming the
fix. Transient failures (API errors reading the parent message)
remain plain `Error`s by design. `executeDispatchRun`
(`.opencode/plugins/review-dispatch/index.ts`, L3182) catches both
kinds at two sites:

- the host-source branch (L3237-3267, RD-FR-037), and
- the advisor fallback inside the tier-based branch
  (L3273-3297, #693).

Both currently map every failure to error code
`host_model_unavailable`, `retryable: true`, and
`sanitizeInvocationError(errorText(error))`. Because `errorText()`
(L114) returns only `err.message`, the structural failure's
`remediation` string is dropped from every recorded form of the
error: the `dispatch_agent_run` result and the
`redispatch_failed_runs` early-exit result entry that repeats
`outcome.result.error` (L7451-7454).

Three defects follow: a structural failure is advertised as
retryable even though retrying cannot fix it; the remediation
string is lost; and no consumer can distinguish the classes
without parsing message prose. A fourth gap is specification:
`retryable` on a pre-execution early exit is never defined. No run
file exists, so `redispatch_failed_runs` (DR-FR-002) has nothing
to target, and consolidation of a session whose included plan
entry early-exited fails closed with "included plan entry
`<agent>/<sequence>` has no run" → `INCONCLUSIVE`.

The classification precedent already exists in this repository:
`uf_delegate_command` (`.opencode/plugins/uf-workflow/index.ts`,
L2563-2573) checks `err instanceof StructuralHostModelError` and
surfaces a non-retryable failure carrying
`<message>. Remediation: <remediation>; retrying will not fix
this.` This change brings the dispatch core to those semantics.

## Goals / Non-Goals

### Goals

- Classify host-resolution failures as structural or transient
  identically on the host-source branch and the advisor fallback,
  through one shared helper rather than mirrored edits.
- Drivers of the split: structural failures surface
  `host_model_structural` / `retryable: false` with the shared
  failure's remediation string preserved; transient failures keep
  today's `host_model_unavailable` / `retryable: true` class.
- State in the spec record that `retryable` on a pre-execution
  early exit means caller re-invocation only: no run file exists,
  so DR-FR-002 redispatch has no target and consolidation fails
  closed with "included plan entry has no run" → `INCONCLUSIVE`.
- Keep the abort-signal `cancelled` path, which takes precedence
  over classification.
- Byte-identical scaffold mirror; test coverage per the 94 percent
  statements floor; CI replication via `make plugin-test` and the
  Go scaffold drift gate.

### Non-Goals

- No `invoke_agent` change: its identical conflation
  (`.opencode/plugins/invoke-agent/index.ts`, L192) is recorded
  for separate follow-up.
- No `uf_delegate_command` change: its uppercase-code rendering
  and existing classification stay as they are.
- No global `errorText()` or `StructuralHostModelError` change:
  `errorText()` is shared by many callers, and the remediation is
  composed only at the dispatch classification seam where the
  shared failure's type is known. The `remediation` field already
  exists on the error.
- No planner, finalizer, input-schema, run-file-shape, prompt
  builder, or consolidation change.
- No change to the retryable classification of any other
  pre-execution early exit (`policy_load_failed`,
  `unknown_agent`, `tier_model_unavailable`, `invalid_invocation`,
  `prompt_file_read_failed`).
- No new `dispatch-agent-run` capability delta: the tool's input
  and tier-default contract is unchanged; the new class appears
  only inside the existing failed-result `error: { code,
  message, retryable }` shape, which RD-FR-038 governs.

## Decisions

### D1. One shared classification helper (chosen)

Add one unexported module-level helper next to
`executeDispatchRun` and call it from both catch sites:

```
function hostResolutionFailureResult(
  error: unknown,
  aborted: boolean,
  provenance: InvocationProvenance,
): InvokeAgentResult {
  if (aborted) {
    return failedResult(
      provenance, "cancelled", "dispatch cancelled", false)
  }
  if (error instanceof StructuralHostModelError) {
    return failedResult(
      provenance,
      "host_model_structural",
      `${error.message}. Remediation: ${error.remediation}; ` +
        "retrying will not fix this.",
      false,
    )
  }
  return failedResult(
    provenance, "host_model_unavailable", errorText(error), true)
}
```

Each catch block becomes
`return earlyExit(hostResolutionFailureResult(error,
context.abort.aborted, earlyProvenance(...)))`, preserving the
existing per-branch provenance (host: no tier; advisor: the tier
label). `failedResult()` already applies
`sanitizeInvocationError()` (L184), so the helper passes raw text
and sanitization stays single-pointed; the existing explicit
`sanitizeInvocationError(errorText(...))` call is redundant with
it and is dropped at this seam.

Rationale: advisor/host symmetry holds by construction, which is
the parent change's own design D1 principle. Two mirrored catch
blocks would be free to drift again — exactly the defect #914
records. `instanceof` also matches the established
`uf_delegate_command` precedent; matching on `error.name` would
be string-fragile across transpilation and realms.

Alternatives rejected:

- **Mirrored edits in both catch blocks.** Preserves today's
  shape but re-creates the drift risk that produced the parent
  D2 note.
- **Discriminate on `error.name === "StructuralHostModelError"`.**
  Avoids importing the class but couples behavior to a mutable
  string property; `instanceof` is the precedent and the class is
  already exported from the shared library.
- **Put classification in `errorText()` or the error class.**
  `errorText()` has many unrelated callers and no remediation
  context; the class cannot know the dispatch result shape.

### D2. Structural mapping: `host_model_structural`, non-retryable

Structural failures (`error instanceof StructuralHostModelError`)
fail with error code `host_model_structural`, `retryable: false`,
and the message
`<error.message>. Remediation: <error.remediation>; retrying will
not fix this.` — the `uf_delegate_command` shape, passed through
the shared sanitizer via `failedResult()`.

Rationale: the code is machine-distinguishable from the transient
class, so consumers branch on `error.code` and `error.retryable`
instead of prose. The separate code also keeps verdict projection
honest: `availabilityError()` (L2136) requires an availability
token (`unavailable`, `availability`, `timeout`, `unsupported`, or
`_not_found`) next to `provider`/`model`/`runtime`;
`host_model_structural` matches none, correctly reading a
structural cause as a configuration problem rather than an
availability blocker. (This path writes no run file, so the
projection is unaffected in practice; the property matters if the
code ever persists.)

Alternatives rejected:

- **Reuse `host_model_unavailable` with `retryable: false`.**
  Callers would still need prose to tell the classes apart, and
  `availabilityError()` would keep classifying it as an
  availability failure.
- **Reuse `invalid_invocation`.** Conflates a runtime host-state
  cause with pre-execution input validation and loses the
  host-model domain.
- **Append the remediation to the existing code's message.** Keeps
  the retryable lie and leaves no structural discriminator.

### D3. Transient mapping unchanged

Any other `resolveHostModel()` failure keeps error code
`host_model_unavailable`, `retryable: true`, and the sanitized
`errorText()` message exactly as today, with no remediation
suffix. The existing transient tests in
`.opencode/test/dispatch-agent-run.test.ts` (advisor L836-875,
host L1003-1035) pin this contract and MUST stay green.

### D4. Abort precedence preserved

The helper checks `aborted` before classification and returns
`cancelled` / `retryable: false` / "dispatch cancelled". An abort
that coincides with a structural error therefore stays
`cancelled`, matching RD-FR-037's scenario and today's behavior.

### D5. Pre-execution early-exit semantics move into the spec

Both classified failures are pre-execution early exits:
`runData` is null, no run file is written, and `persisted` is
false. The spec records what `retryable` means there:

- `retryable: true` (transient) means the caller MAY re-invoke the
  dispatch after the lookup recovers; `retryable: false`
  (structural) means re-invocation cannot succeed until the
  underlying cause is fixed.
- Neither class creates a DR-FR-002 redispatch target: with no run
  file, `redispatch_failed_runs` has nothing to target (a named
  agent reports `not_found`; otherwise the agent is absent from
  scoping).
- Consolidation of a session whose included plan entry early
  exited fails closed with `included plan entry
  <agent>/<sequence> has no run` → `INCONCLUSIVE`.
- `redispatch_failed_runs`' own re-run early exit is distinct: it
  restores the prior run file and reports outcome `early_exit`
  (with the classified `error`), returning the session to its
  pre-attempt state (L7447-7454).

Rationale: operators currently infer this from behavior; #914
asks for it in the record.

### D6. No `dispatch-agent-run` capability delta (chosen)

The `dispatch_agent_run Tool` requirement (shared-agent-executor
delta, MODIFIED by `host-fallback-provenance`) governs parameters,
tier default, and the general failed-result shape; it does not
enumerate host-resolution error codes. RD-FR-038 supplies the
behavior inside the existing `error: { code, message, retryable }`
shape, so no dispatch-agent-run requirement text changes. Adding a
delta would duplicate the same contract across capabilities.

### D7. Cross-delta supersession and archive ordering

Two live requirement texts are falsified and therefore MODIFIED in
the review-dispatch delta, each with a Previously paragraph:

- **Host-Source Model Resolution (RD-FR-037)** — added by
  `host-fallback-provenance` — whose failure clause currently
  routes every failure to `host_model_unavailable` /
  `retryable: true`.
- **Advisor Host-Model Fallback** — MODIFIED by
  `host-fallback-provenance` — whose "Host resolution failure"
  scenario carries the same unconditional retryable contract.

Both MODIFY entries apply against the materialized predecessor
text, so this change MUST archive after
`host-fallback-provenance` (which itself archives after
`advisor-host-model-fallback` and `shared-agent-executor`). No
older change is edited in place. Strategy mirrors the parent's D6.

### D8. Test strategy

Extend the existing suites; no new harness. Both branches are
driven through `dispatchAgentRun` and the shared core, using the
existing doubles:

- `.opencode/test/dispatch-agent-run.test.ts` — reuse the
  `hostLookupClient({ message })` shaper (L932) and the empty
  `standard` profile matrix (advisor) for structural shapes: parent
  message role not `assistant`, missing `providerID`/`modelID`,
  and a present-but-invalid `variant`. Cover the host branch
  (`source: "host"`) and the advisor branch (`tier: "standard"`,
  `source: "advisor"`) to prove the split is symmetric by
  construction. Keep the existing transient test and add its
  structural counterpart. Assert code, `retryable`, remediation
  presence and redaction, and an empty run-file listing.
- `.opencode/test/dispatch-recovery.test.ts` — seed a failed
  host-source run, make host resolution fail structurally on the
  re-run, and assert outcome `early_exit`, the classified
  `error` in the result entry, and restoration of the prior run
  file (the L1207-1264 pattern).
- `.opencode/test/consolidate-dispatch.test.ts` — a direct
  dispatch whose included plan entry early-exited structurally
  produces no run; consolidation MUST fail closed with
  `included plan entry <agent>/<sequence> has no run` and an
  `INCONCLUSIVE` verdict.

### D9. Scaffold mirror and drift gate

Mirror the live plugin edit byte-identical into
`internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`.
`TestEmbeddedAssets_MatchSource`
(`internal/scaffold/scaffold_test.go`) is the enforcement; no
scaffold-side code change beyond the copy.

### D10. Documentation

`docs/configuration.md` "Host-fallback model resolution"
(L296-312) currently documents only the retryable
`host_model_unavailable` class; it gains the structural class, the
remediation string, and the caller-re-invocation-only meaning of
`retryable`. CHANGELOG.md gains an entry. The change is
user-facing (operators and calling agents consume the error
class), so the Cross-Repo Documentation rule applies: an in-repo
documentation issue and an unbound-force/website issue are filed
before PR merge.

## Risks / Trade-offs

- **Observable behavior change for structural failures.** A
  consumer that branch-matched `host_model_unavailable` for every
  host-resolution failure will now see `host_model_structural`
  with `retryable: false`. This is the intended fix; the CHANGELOG
  and docs name both classes, and the transient class is unchanged.
- **Sanitization could truncate the remediation suffix.**
  `failedResult()` truncates to `MAX_ERROR_LENGTH`; structural
  messages are short constants, so the suffix survives, but a
  future long structural message could clip it. Accepted; the
  alternative (unsanitized or unbounded text) is worse.
- **Coverage floor risk.** The 94 percent statements floor is
  enforced by `make plugin-test`. The helper and two call sites
  are small and covered by the mirrored branch tests. If the floor
  fails, the gatekeeping rule applies: report and stop; never
  adjust `.opencode/vitest.config.ts`.
- **Cross-realm `instanceof`.** The check assumes one module graph,
  which holds for both the live plugin and its scaffold copy.
  Tests import the same module. No duplicate-package risk exists.
- **Documentation drift.** The website copy of the failure
  semantics can lag; the documentation issue is filed in this
  change so the gap is tracked.
