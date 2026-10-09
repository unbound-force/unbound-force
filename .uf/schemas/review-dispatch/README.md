# Review Dispatch Schema

The `review-dispatch` payload records model-level execution and
provenance for one Divisor workflow. It supports audit and diagnosis;
it does not replace the canonical downstream review decision.

## Producer and Consumers

| Role | Hero | Relationship |
| --- | --- | --- |
| Producer | The Divisor | Emits dispatch provenance |
| Consumer | The Divisor | Diagnoses and reruns dispatches |
| Consumer | Mx F | Observes coverage, cost, and failures |

The Divisor emits this artifact in addition to `review-verdict` 2.0.0.
Mx F, Cobalt-Crush, and Muti-Mind continue to consume
`review-verdict` as the canonical decision artifact. Consumers MUST
NOT infer approval from `review-dispatch` alone.

## Contract

The payload schema is draft 2020-12 and closed at every object level.
It defines command and mode identity, immutable input context, the
version 1 plan, run provenance, findings, advisories, coverage, native
workflow result, generic verdict, and seven terminal run counts.

JSON Schema owns structure, types, enums, bounds, and nullability.
`finalize_review_dispatch` owns cross-field semantics, including run
count arithmetic, final-only states, timestamp ordering, model identity
agreement, native-result projection, and persistence.

## Samples

- `samples/sample-review-dispatch.json` is a successful PR review.
- `samples/valid-local-review-dispatch.json` is a local no-success run.
- `samples/valid-issue-review-dispatch.json` is issue triage.
- `../samples/sample-review-dispatch-envelope.json` is the complete
  Hero Interface Contract envelope.

Files beginning with `invalid-` exercise rejected structure, enums,
nullability, bounds, discriminators, and unknown properties.

## Version History

| Version | Date | Changes |
| --- | --- | --- |
| 1.0.0 | 2026-10-01 | Initial observability contract |
