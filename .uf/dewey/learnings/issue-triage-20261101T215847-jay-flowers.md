---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:58:47Z
identity: issue-triage-20261101T215847-jay-flowers
tier: draft
---

Unbound-force triage pattern: repo-local governance packs carry self-declared posture in their own headers — determinism.md and meta-prompt.md both open with "repo-local governance pack (not a scaffold asset)", and meta-prompt.md's Custom Rules section states it "ships no -custom.md companion" per the determinism precedent. An issue proposing to deploy such a pack via uf init (alwaysDeployedPacks entry plus a *-custom.md template, as issue #889 does for meta-prompt) is reversing that declared posture. Treat the issue as VALID when its acceptance criteria are unambiguous and testable — the reversal is the issue's explicit intent — but flag in the triage reasoning that the implementing change must update the pack's self-description, or the deployed pack ships stale posture statements.
