---
tag: dewey-learning-pr-policy
author: jay-flowers
category: gotcha
created_at: 2026-09-30T02:10:37Z
identity: dewey-learning-pr-policy-20260930T021037-jay-flowers
tier: draft
---

Scaffold drift sync is critical when modifying any file under .opencode/ that has an embedded copy under internal/scaffold/assets/. The CI job TestEmbeddedAssets_MatchSource fails on any byte-level divergence. The safe pattern is: edit the canonical file, then `cp canonical embedded` followed by `diff canonical embedded` to confirm identity. For uf.unleash.md specifically, both .opencode/commands/uf.unleash.md and internal/scaffold/assets/opencode/commands/uf.unleash.md must remain byte-identical.
