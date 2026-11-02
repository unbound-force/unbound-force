---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:22:43Z
identity: issue-triage-20261101T222243-jay-flowers
tier: draft
---

Deferred-scope issue triage pattern: when an issue cites lineage (parent acceptance criterion, delivering PR, design.md Non-Goals) but the cited artifacts are absent from the current working tree, search the indexed tracker content for the delivering PR before downgrading the verdict. An open delivering PR explains the artifact absence — the change directory and package live on the PR branch, not the default-branch checkout — and the follow-up issue remains VALID as the queued tracker for the deferred criterion. Cross-check the PR's stated contract against the issue's claims (exit-code taxonomy, flags, output streams) to confirm the premise. Caveat: tracker indexes lag real time (the github-unbound-force index ran about four days before the issue was filed), so index absence is weak evidence of nonexistence, never proof.
