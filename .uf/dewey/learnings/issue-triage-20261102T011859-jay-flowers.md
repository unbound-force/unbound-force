---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:18:59Z
identity: issue-triage-20261102T011859-jay-flowers
tier: draft
---

Duplication-removal issues go stale between filing and implementation in fast-moving repos. unbound-force issue #769 (filed 2026-10-19) listed seven .opencode/test suites holding local withBun copies, but by triage time (2026-11-01) two of the listed suites (dispatch-status, dispatch-review-runs) had been migrated to the shared helpers/bun-runtime.ts module by a later unrelated change, an unlisted suite (read-diff-chunk.test.ts) had gained a local copy, and the review-dispatch plugin had grown from 5,992 to 8,634 lines. The CS-004 violation persisted, so the issue stayed valid, but its inventory was wrong in both directions. Triage and implementation of duplication issues must re-derive the occurrence list from the current tree (grep the duplicated symbol) rather than executing the filing-time list.
