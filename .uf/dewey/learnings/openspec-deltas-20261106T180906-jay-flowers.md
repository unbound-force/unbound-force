---
tag: openspec-deltas
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:09:06Z
identity: openspec-deltas-20261106T180906-jay-flowers
tier: draft
---

When an OpenSpec change must modify requirement text whose canonical form will only exist after unarchived sibling changes archive, put the reconciled text in the siblings' deltas (citing file:line for each edit) and record the archive ordering in the new change's design — a MODIFIED requirement in the new change cannot match canonical text that does not exist yet. Precedent: stop-shipping-plugin-lock's Cross-Change Reconciliation section amending ts-plugin-coverage-gate, init-override-stubs-gitignore-sync, and fix-triage-issue-dispatch-docs in place instead of shipping its own ci delta.
