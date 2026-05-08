---
tag: scaffold
category: gotcha
created_at: 2026-04-28T19:02:00Z
identity: scaffold-1
tier: draft
---

When cherry-picking commits that include scaffold asset syncs across branches, the OpenSpec template files (openspec/schemas/unbound-force/templates/*.md) frequently end up with duplicate scaffold markers because both the source and target branches independently added markers. The fix is to run a dedup pass after cherry-pick: for each .md file with 2+ markers, keep only the first. Both the live files AND their scaffold asset copies must be deduped and synced, otherwise TestEmbeddedAssets_MatchSource and TestEmbeddedAssets_SingleMarker will fail. This happened on both the gateway-token-refresh-fix and sandbox-uid-mapping branches.
