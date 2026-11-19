---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:26:02Z
identity: review-dispatch-20261110T212602-jay-flowers
tier: draft
---

A display-truncation cap applied to data that consumers persist verbatim is a data-damage bug, not a display concern. When a bounded tool-result preview feeds a persistence path (lesson proposals → dewey_store_learning), bound it whole-or-absent instead: copy every entry field verbatim, drop over-budget entries whole (trailing first), and flag the drop with a count-level marker (proposals_truncated). A per-entry truncated:true marker does not protect consumers that ignore it — PR #955 stored 11 of 12 council learnings cut mid-word at the old 512-character cap despite the marker and the store-from-artifact prose contract. Removing the cut at the source makes every retrieval mode safe to harvest; the RD-FR-058 change in unbound-force PR #1004 is the shipped precedent (projectSummaryProposal copies verbatim; buildSummaryRetrieval and buildCompactFinalization each enforce a post-assembly 40,960-byte budget by dropping whole entries).
