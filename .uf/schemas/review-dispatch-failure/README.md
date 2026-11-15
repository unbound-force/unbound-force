# Review Dispatch Failure Schema

The `review-dispatch-failure` payload records one failed
review-dispatch finalization (RD-FR-066, issue #1039): the
complete, never-truncated errors of the failure, the dispatch
correlation id when one was UUID-usable (null otherwise), the
command and mode when the failed input knew them, and the
operation verdict the finalizer computed before failing. The
record exists so a failed dispatch stays diagnosable after the
fact.

## Producer and Consumers

| Role | Hero | Relationship |
| --- | --- | --- |
| Producer | The Divisor | Best-effort record on every failed finalization |
| Consumer | The Divisor | Reads the complete errors through `read_dispatch_artifact` |
| Consumer | Mx F | Diagnoses why a dispatch failed |

The Divisor writes the payload inside a Hero Interface
Contract envelope with `artifact_type:
"review-dispatch-failure"` at
`.uf/artifacts/dispatch/<correlation-id>.failure.json` —
keyed by the dispatch's correlation id when UUID-usable, a
freshly minted one otherwise. A dispatch that finalizes
successfully leaves no record. Consumers MUST NOT infer any
review decision from a failure record: the operation verdict
it carries is the fail-closed outcome of the failed
finalization, not a review of the code.

## Contract

The payload schema is draft 2020-12 and closed: `status` is
fixed to `"failed"` (the record exists only on failure
paths), unknown keys are rejected so a tampered record fails
closed when read back, and `errors` carries the finalizer's
complete array verbatim. Sanity bounds sit above the natural
error-composition ceilings — 131,072 characters per entry and
65,536 entries — and fail closed: a composition beyond them
leaves no record rather than a silently damaged one.

JSON Schema owns structure, types, enums, const, bounds, and
nullability. The envelope builder owns fail-closed payload
validation and best-effort provenance; the finalizer owns the
three failure-path triggers (structural rejection, semantic
rejection, persistence failure) and the never-mask
degradation when the record write itself fails.

## Samples

- `samples/sample-review-dispatch-failure.json` is a semantic
  rejection: a zero-success payload claimed an approving
  verdict, and the cross-checks failed.
- `samples/valid-null-correlation-id.json` is a structural
  rejection of a garbage input: no usable dispatch id (null),
  and command and mode stay absent.

Files beginning with `invalid-` exercise one rejected
contract each:

- `invalid-correlation-id.json` — non-UUID, non-null
  `dispatch_correlation_id`
- `invalid-command.json` — `command` outside its enum
- `invalid-mode.json` — `mode` outside its enum
- `invalid-status.json` — `status` is the const `"failed"`
- `invalid-operation-verdict.json` — `operation_verdict`
  outside its enum
- `invalid-empty-errors.json` — empty `errors` array
  (`minItems` 1)
- `invalid-empty-error-string.json` — empty error string
  (`minLength` 1)
- `invalid-error-entry-too-long.json` — 131,073-character
  entry (`maxLength` 131,072)
- `invalid-too-many-errors.json` — 65,537 entries
  (`maxItems` 65,536)
- `invalid-missing-required.json` — missing the required
  `errors` key
- `invalid-extra-property.json` — unknown payload key

## Version History

| Version | Date | Changes |
| --- | --- | --- |
| 1.0.0 | 2026-11-15 | Initial failure-record contract (RD-FR-066) |
