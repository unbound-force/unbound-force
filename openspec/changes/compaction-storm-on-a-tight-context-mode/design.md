## Context

The `/uf.unleash` pipeline delegates work to child agents
via `uf_delegate_command` / `invoke_agent` using
`agent-executor.ts` + `delegation-budget.ts` and the
`model-matrix` selection layer. The 2026-11-14 incident
showed three compounding faults:

1. **Headroom-blind selection** — `resolveSelection`
   chose `Muse-Spark-Open-816B-A42B-NVFP4` because its
   declared `limit.context - limit.output` (~98k) passed
   existing validation, yet with the opencode prefix
   (~76.5k cached) only ~20k remained for work. Every
   file read compacted.
2. **Storm invisibility** — `messageFingerprint` /
   `probeSubtreeActivity` treated compaction summaries as
   progress (`finish_reason`, `finish_summary`), so the
   idle watchdog never fired despite ~150 compactions in
   110 min (~every 30s) and an 18-min
   `dispatch_review_runs` stall.
3. **False success** — a child whose final text was a
   compaction summary was reported `status: success` with
   `checkpoint_recorded: false`, indistinguishable from a
   legitimate gate pause by the orchestrator.

This change closes those gaps in the current delegation
stack (`model-matrix.ts`, `delegation-budget.ts`,
`agent-executor.ts`, `uf-workflow/index.ts`) and is
bounded to delegation; a future orchestrator change can
consume the new failure code/resume envelope.

## Goals / Non-Goals

### Goals

- Reject or penalize models whose usable context cannot
  fit the command's base footprint plus working memory;
  record why in `selection_fallback_reason`.
- Detect N compactions without non-read progress and fail
  fast with `context_thrash` + `resume` envelope; surface
  to orchestrator as retryable.
- Never report compaction-summary / synthetic-prompt
  final text as `success`.
- Keep all new behavior deterministic, tested, and
  observable (audit fields).

### Non-Goals

- Changing orchestrator resume orchestration (follow-up
  change consumes `context_thrash`/`resume`).
- Retuning model limits, classifier prompts, or
  compaction summarization itself.
- New cross-hero dependencies; heroes remain
  independently installable.

## Decisions

### D1 — Headroom gate in selection (model-matrix.ts)

**Decision:** After the existing precedence chain,
compute `usable = limit.context - limit.output` (when
declared) and compare to `footprint + margin`.
`footprint` = cached prefix + command body estimated from
on-disk `uf-workflow` / `unleash` content; `margin` =
minimum working memory (default 24k, configurable via
`model-matrix.yaml`). If `usable < footprint + margin`,
the candidate is rejected (or heavily penalized) and the
matrix falls to the next eligible selection. The rejection
chain is concatenated into `selection_fallback_reason`
(e.g. `classifier model enmaas/... usable 98k < required
100k (76.5k footprint + 24k margin) — fell back to ...`).

**Purity and path safety:** the footprint is computed by
the caller (the `uf-workflow` plugin) and supplied via
`SelectionInputs`; `resolveSelection` itself performs no
filesystem access and stays pure. The footprint estimator
constrains reads to repo-root command paths joined via
path-join (rejecting traversal and symlink escapes); on
any I/O failure it falls back to a documented,
conservative default footprint (fail-closed: the default
is sized so a read failure cannot understate the
footprint) and records the fallback in the selection
audit.

**Terminal behavior:** when every candidate is
headroom-rejected (or declares inconsistent limits),
selection falls back to the host model with the full
rejection chain in `selection_fallback_reason`; it never
silently proceeds with a rejected candidate and never
stalls without a result. A candidate declaring
`limit.output >= limit.context` is rejected with a reason
naming the inconsistency; a candidate with no declared
limits is not headroom-rejected.

**Alternatives:** (a) Hard-reject only at spawn time —
rejected: wasted classifier work and audit gap. (b)
Token-count exact measurement — rejected: needs provider
tokenizer; estimate + margin is sufficient and
deterministic.

**Constitution:** II Composability — matrix stays pure
(inputs: limits, caller-supplied footprint); III
Observable Quality — reason recorded; VI Deterministic
Execution — arithmetic, not judgment.

### D2 — Compaction-storm detection
(delegation-budget.ts + agent-executor.ts)

**Decision:** `ProbeState` tracks
`compactionsSinceProgress` and `nonReadProgressCount`. A
message that structurally matches a compaction marker
(user-role compaction-summary wrapper or synthetic
continuation prompt; see D3 for matcher discipline)
increments `compactionsSinceProgress`; any successful
non-read tool completion resets it. When `N` consecutive
compaction increments occur without intervening non-read
progress (default N=3, configurable), the executor aborts
the child and returns a new `InvokeAgentResult` failure
with `code: "context_thrash"` (retryable), `error` naming
the window, and the existing `timeout.details` + new
`resume` envelope (reusing the `DelegateResumeState`
shape) carrying durable state (`project_path`,
`agent_name`, pending `checkpoint_step`).

`messageFingerprint` is amended so compaction markers DO
NOT count as progress; only successfully completed
non-read tool calls, file writes, and checkpoint records
do (reads and failed calls never count).
`probeSubtreeActivity` exposes `compactionCount` for the
watchdog.

**Alternatives:** (a) Extend idle timeout — rejected:
conflates wall-clock stalls with compaction loops; storm
needs structural detection. (b) Count total compactions —
rejected: a healthy long run may compact occasionally;
the "since progress" window is the signal.

**Constitution:** IV Testability — deterministic counter
via injected `now`; I Autonomous Collaboration — `resume`
envelope is artifact-based.

### D3 — Final-text success guard (agent-executor.ts)

**Decision:** Post-execution classifier inspects the
child's final text before mapping to
`OperationStatus.SUCCESS`. Matchers are structural:
anchored, case-insensitive, and length-bounded against
the known wrapper shapes — a compaction-summary wrapper
(opencode's system-reminder compaction structure) or a
known synthetic continuation/handoff prompt prefix —
never bare substrings such as the word "compaction" that
can appear in legitimate prose. On a match the result is
reclassified to `context_thrash` (if the storm threshold
was met) or a generic retryable failure with
`code: "synthetic_final_text"`. The original text is
preserved in `error` for debugging, truncated to a
documented maximum length and passed through
`sanitizeInvocationError` so no cached-prefix content or
secrets leak; no `success` envelope is emitted.

**Constitution:** III Observable Quality — false success
eliminated; V Security by Default — sanitized preserved
text; VI Deterministic Execution — classification is
mechanical, covered by unit tests including
negative controls (legitimate prose mentioning
compaction stays success).

### D4 — Configurability

`model-matrix.yaml` gains optional
`selection.footprint_margin_tokens` and
`delegation.storm_threshold` (defaults 24k / 3). Each knob
is validated on load: an integer within a documented range
(margin 1k–256k tokens; threshold 2–10). A missing value
falls back to the default. An invalid value (wrong type,
out of range, unparsable) is replaced by the default and
the substitution is logged as a warning event naming the
key, the invalid value, and the default applied, so a
misconfigured knob cannot silently disable the gate or
detector. Malformed YAML fails closed as today. Docs
updated via tool-owned asset `docs/configuration.md`
scaffold.

**Constitution:** V Security by Default — fail-closed
validation with logged substitutions; III Observable
Quality — substitution events observable.

## Contracts

New runtime surfaces introduced by this change (consumed
by `uf_delegate_command` and future orchestrator changes):

- `InvokeAgentResult` additions: failure codes
  `context_thrash` (retryable; storm threshold reached)
  and `synthetic_final_text` (retryable; final-text
  guard). Both carry `error` details naming the storm
  window or matched structure and reuse the existing
  `timeout.details` shape.
- `ProbeState` additions (delegation-budget.ts):
  `compactionsSinceProgress`, `nonReadProgressCount`,
  `compactionCount`, exposed through
  `probeSubtreeActivity`.
- `SelectionInputs` addition: caller-computed `footprint`
  (tokens). Selection audit addition:
  `selection_fallback_reason` headroom entries naming the
  rejected model, usable, required, footprint, margin,
  and the fallback taken.
- Resume envelope: `context_thrash` results carry durable
  state (project path, agent name, last recorded
  checkpoint step) in the existing resume envelope
  pattern (DT-FR-004), which `uf_delegate_command`
  threads for command-path delegations as `resume` with
  `command`, `key`, `last_checkpoint_step`,
  `completed_steps`.
- Config keys: `selection.footprint_margin_tokens`,
  `delegation.storm_threshold` (see D4).

These surfaces are the seams verified by the
seam-verification coverage in tasks.md (registration
surface, documented output shape, write→read round-trip).

## Risks / Trade-offs

- **Headroom estimate inaccuracy** — estimated footprint
  may be off by a few k; margin absorbs error. If still
  wrong, fallback_reason makes mis-selection debuggable
  and threshold is tunable.
- **Storm threshold tuning** — too low (N=2) could abort
  legitimate compaction-heavy but progressing runs; too
  high (N>=5) delays abort. Default N=3 chosen from
  incident (150 compactions, zero non-read progress);
  configurable and covered by tests for N boundaries.
- **Final-text pattern drift** — synthetic prompts may
  change wording; classifiers match structural wrappers,
  not wording, and the fallback is that the storm detector
  still aborts, so the guard is defense-in-depth.
- **Scope creep into orchestrator** — this change
  deliberately stops at executor failure envelope;
  orchestrator resume behavior is a separate change to
  avoid coupling.

## References

- Incident: `dewey#203`, issue #1079 (companion
  gate-integrity track).
- Prior fixes: #930/#950 idle timeout, #380 state loss.
- Files: `.opencode/lib/model-matrix.ts`,
  `.opencode/lib/delegation-budget.ts`,
  `.opencode/lib/agent-executor.ts`,
  `.opencode/plugins/uf-workflow/index.ts`,
  `.uf/model-matrix.yaml`, `docs/configuration.md`.
