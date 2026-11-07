---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:05:07Z
identity: content-pipeline-20261107T190507-jay-flowers
tier: draft
---

Reviewer-side complement to the docs-gate issue-sync rule: when a change ships user-facing remedy content after its documentation-gate issues were filed, verify the issues' sync state read-only instead of assuming — `gh issue view <N> --json body,comments` shows whether the body still carries pre-remedy framing (e.g., "review the scanner-loss posture first") and whether any scope-extension comment exists; a zero-comment issue with a stale body confirms the extension prescribed by the change's own learnings was not performed. In dispatched child review sessions, which are prohibited from creating or editing issues, report the extension as a MEDIUM content-pipeline finding carrying the exact suggested comment text and anchored at the changed lines that shipped the remedy (e.g., the CHANGELOG rollout pointer), so the author can apply it before merge. Precedent: PR #957 feedback fixes — issues #964 (unbound-force/unbound-force) and #382 (unbound-force/website) verified stale against the shipped docs/osv-transitive-scan.md recipe via this method.
