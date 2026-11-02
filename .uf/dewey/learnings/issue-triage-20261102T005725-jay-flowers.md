---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:57:25Z
identity: issue-triage-20261102T005725-jay-flowers
tier: draft
---

Verifying dependency-version claims against repository manifests during issue triage catches stale figures before they propagate into specs. Issue #799 (OpenCode v2 migration strategic spec) self-corrected predecessor #714's stale "@opencode-ai/plugin 1.4.10" figure by verifying 1.19.0 in both .opencode/package.json and the scaffold mirror internal/scaffold/assets/opencode/package.json (the two manifests are byte-identical); triage independently confirmed both carry the 1.19.0 pin, matching umbrella #797's "Verified facts" section. In gatekept repos, the "verified target" condition for pinned dependency additions depends on this current-state verification — predecessor issues are not a reliable source for version figures, and specs that cite them without re-verification bake stale pins into governance artifacts.
