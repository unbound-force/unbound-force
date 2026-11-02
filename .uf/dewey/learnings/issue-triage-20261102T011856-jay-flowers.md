---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:18:56Z
identity: issue-triage-20261102T011856-jay-flowers
tier: draft
---

Deferred review-finding follow-up issues must be re-verified against current repository state at triage time, not trusted from their filing-date snapshot. In unbound-force issue #769 (filed 2026-10-19), the shared withBun helper (.opencode/test/helpers/bun-runtime.ts) and the aggregate-dispatch fixture factory (helpers/aggregate-dispatch-fixtures.ts, fix commit 12c5517 for #794) were created by later fixes after filing, so 2 of the 7 named suites had already migrated — narrowing the task from extraction to migration — while a NEW copy site accreted (read-diff-chunk.test.ts) and the plugin size claim grew stale in the strengthening direction (5,992 → 8,634 lines). To establish the real CS-004 copy count, grep for the definition pattern ("function X"), not usages — usage greps truncate at high match counts and conflate migrated importers with local copiers.
