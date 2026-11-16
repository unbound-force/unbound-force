---
tag: uf-unleash-environment
author: jay-flowers
category: gotcha
created_at: 2026-11-16T09:55:01Z
identity: uf-unleash-environment-20261116T095501-jay-flowers
tier: draft
---

Running /uf.unleash in this checkout (2026-11-16): dewey_store_learning fails with "store_learning requires persistent storage. Configure --vault with a .uf/dewey/ directory" even though .uf/dewey exists on disk — the MCP server is not vault-configured, so dispatch lesson proposals and retrospective learnings must be captured by hand: a session note for in-flight stranded-proposal tracking, then learning files written directly to .uf/dewey/learnings/ following the vault frontmatter convention (tag/author/category/created_at/identity/tier: draft) so the daily chore/dewey-learnings-* PR still picks them up. Two dispatch-quality observations from the same run: divisor-adversary died with a context_thrash abort on review iteration 1 (retryable — three consecutive compactions with no substantive progress, the storm detector working as designed) and produced a clean run on the re-dispatch, so a single context_thrash failure does not poison a verdict while other reviewers succeed and the excluded orphaned findings are simply re-harvested on the re-run; and empty_coverage_attestation warnings appeared for reviewers whose final text omitted the Files Reviewed attestation — advisory quality signals only, never a finding or verdict change. Tags: opsx/per-task-delegation-budgets, 2026-11-16, gotcha.
