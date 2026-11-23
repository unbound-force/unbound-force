---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:49:30Z
identity: uf-unleash-20261116T014930-jay-flowers
tier: draft
---

In /uf.unleash code review over an uncommitted-local change (branch opsx/feedback-triage-artifact-tool, 2026-11-16), capture_review_diff assembles tracked modifications plus untracked files only under the feature dir — untracked implementation files outside it (a new .opencode/lib module, its test file, and the internal/scaffold/assets mirror) were silently absent from the first captured diff (19 files instead of 22, missing the 1595-line core deliverable). Recording intent-to-add entries (git add -N <paths>) before capture makes new files diff-visible without committing anything, keeping the tree uncommitted for /uf.finale. Always verify the captured files list actually contains every new implementation file before dispatching the review council.
