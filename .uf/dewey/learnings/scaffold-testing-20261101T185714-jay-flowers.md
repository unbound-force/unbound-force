---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T18:57:14Z
identity: scaffold-testing-20261101T185714-jay-flowers
tier: draft
---

To sanity-check that an embed-FS guard test in internal/scaffold fails correctly: temporarily move the embedded asset file out of internal/scaffold/assets/ (e.g. mv to a temp dir), run only the guard with go test -race -count=1 -run 'TestName' ./internal/scaffold/ (go test rebuilds the package and re-embeds the changed asset set, so the guard sees the file as missing; -run scoping avoids the cascade of unrelated drift/manifest tests that also fail when an embedded file disappears), then move the file back and verify no residue with git status plus cmp against the canonical source. Used for TestSCFR009_DeploymentCompletenessGuard (model-matrix-scaffold-deployment task 4.4): removing the meta-prompt mirror made the guard fail naming internal/scaffold/assets/opencode/uf/packs/meta-prompt.md, and restoring it returned the suite to green with the mirror byte-identical.
