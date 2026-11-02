---
tag: scaffold-assets
author: jay-flowers
category: reference
created_at: 2026-10-31T02:51:35Z
identity: scaffold-assets-20261031T025135-jay-flowers
tier: draft
---

Verifying cmd/unbound-force/main_test.go embedded-asset arithmetic for scaffold-sync PRs: total embedded assets = len(expectedAssetPaths) + len(nonDeployedAssetPaths) from internal/scaffold/scaffold_test.go (base of PR #882: 77 + 1 = 78; head: 79 after adding one lib asset). The 4 activation-gated assets are exactly the four plugin index.ts files (isActivationGatedAsset, internal/scaffold/scaffold.go:51-52), so the default-deployed count = total − 1 (devcontainer, non-deployed) − 4 (gated) and the asserted range is [deployed, deployed+4]. //go:embed assets (scaffold.go:31, no all: prefix) excludes dot-files, so stray .DS_Store files under internal/scaffold/assets never inflate the count. Counting tip: the glob tool returns 'No files found' for assets/openspec/** patterns even when the files exist — read the directories directly or count expectedAssetPaths entries instead. PR #882's 76→79 count correction was verified this way as a repair, not a defect: the old '76 total' comment was stale (actual base count 78) and the old range [71-75] was latently under-bounded (would fail with 3+ activation probes); the new range [74-78] is exactly correct.
