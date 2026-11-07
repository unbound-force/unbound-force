---
tag: openspec-deltas
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:57:58Z
identity: openspec-deltas-20261105T145758-jay-flowers
tier: draft
---

When an OpenSpec MODIFIED-requirement rewrite narrows an accumulated contract, diff the new clause list against the predecessor's full text: a rewrite that carries forward part of a compound MUST (exact version pins without the co-located license clause) silently drops contract surface unless the delta carries an MP-013 Previously paragraph. Deleted verification code marks the clauses at risk — each assertion deleted with a mechanism (license checks deleted alongside lock assertions in PR #957's scaffold_test.go) corresponds to a contract clause that must either carry forward on a replacement path, move to another requirement, or be recorded as a deliberate removal in design.md. The PR #957 contrast is instructive: the integrity-hash loss was recorded in proposal.md's Impact while the co-deleted license-clause loss was not, and only the unrecorded one became maintainer feedback.
