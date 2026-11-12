---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:15:20Z
identity: review-dispatch-20261110T231520-jay-flowers
tier: draft
---

A lossy preview of durable data is a latent corruption path even when a prose contract forbids consuming it. consolidate_dispatch cut lesson proposals to 512 characters in its summary and compact-finalization previews while the artifact kept full text, and commands were told never to store lessons from the preview (RD-FR-056); PR #955 stored from the preview anyway, damaging 11 of 12 learnings mid-word. RD-FR-058 removed the cut at the source: every proposal field is copied verbatim, and an entry that cannot fit the 40,960-byte result budget is dropped whole and counted by proposals_truncated, so no retrieval mode can hand a consumer a partial body. Pattern: when a bounded display view of persisted data exists, enforce whole-or-absent semantics in the projection itself — a `truncated: true` marker labels damaged data, it does not prevent storing it, because prose contracts do not enforce themselves.
