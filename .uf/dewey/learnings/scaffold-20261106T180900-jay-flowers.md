---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:09:00Z
identity: scaffold-20261106T180900-jay-flowers
tier: draft
---

When an embedded scaffold asset becomes a derived artifact (canonical source minus a block — e.g. the shipped .opencode/package.json is the repository manifest without devDependencies), the derivation needs three synchronized layers plus an updated remediation: (1) a test-side derivation helper for drift expectations (shippedPluginManifest), (2) a production-side re-insertion for refresh (preservePluginDevDependencies), (3) a round-trip test against the live repository artifact asserting preserve(shipped, repo) == repo byte-for-byte — the load-bearing guard, because layout-sensitive byte anchors in either layer fail silently — and (4) the drift test's failure message must change from "cp canonical → asset" to the derivation command, otherwise the first remediation a developer follows reintroduces the stripped block and fails a different assertion. Verified on unbound-force stop-shipping-plugin-lock (PR #957).
