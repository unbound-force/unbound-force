---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:58:52Z
identity: issue-triage-20261101T215852-jay-flowers
tier: draft
---

Governance packs in the unbound-force meta repo carry self-describing design stances in-file that later issues can silently reverse. The meta-prompt convention pack (.opencode/uf/packs/meta-prompt.md, the issue #888 deliverable) documents three stances: a header calling it a "repo-local governance pack (not a scaffold asset)", deferral of determinism rules to determinism.md, and a Custom Rules note that it "ships no -custom.md companion" because project-specific rules are incorporated by revising the pack through a governance change. Issue #889 (slice 2) proposes deploying that pack via uf init as an alwaysDeployedPacks entry with a meta-prompt-custom.md initial-deployment template, reversing the first and third stances without acknowledging them. Triage practice: when an issue changes a file's deployment or distribution status, cross-check the target file's own header and Custom Rules notes for documented design decisions the change would reverse, and require the implementing change to revise those in-file statements in the same increment so the file's self-description never drifts from its deployment reality. determinism.md carries the same repo-local header and marks the boundary between repo-local and scaffolded governance packs.
