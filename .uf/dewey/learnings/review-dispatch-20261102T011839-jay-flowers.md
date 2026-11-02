---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:18:39Z
identity: review-dispatch-20261102T011839-jay-flowers
tier: draft
---

Archive-then-replace recovery flows must gate their success outcome on an explicit persistence signal, never on a non-null executor return. When the executor's run-file write is best-effort (catch-and-continue), a silent write failure after the prior record has been archived leaves the plan entry with no active record while the caller reports success with an id that was never durably written. Shipped fix pattern (unbound-force review-dispatch, DR-FR-002 / issue #771): the shared executor returns a `persisted` boolean set only after a write-temp-then-rename completes; the recovery tool checks `!outcome.persisted`, restores the archived prior file to its active path, and reports an error outcome instead (operator-intervention-required when the restore also fails). Regression-test injection technique: chmod the session directory read-only from inside the child-prompt hook after the archive step, so the replacement write fails deterministically and the test can assert both the reported outcome and the on-disk stranded state (no active run file, prior run preserved under superseded/).
