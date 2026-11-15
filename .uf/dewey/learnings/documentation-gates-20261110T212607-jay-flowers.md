---
tag: documentation-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:26:07Z
identity: documentation-gates-20261110T212607-jay-flowers
tier: draft
---

Unbound Force enforces a two-filing documentation gate for user-facing changes: a docs-labeled issue in the current repo (AGENTS.md Documentation gate) plus a tracking issue in unbound-force/website (constitution Cross-Repo Documentation); the CHANGELOG entry is a separate third assessment and satisfies neither. Precedent: divisor-curator's HIGH finding on PR #1024 produced #1034 and website#405, filed the same minute. Review signals that the gate is open: the change's tasks.md documentation phase stops at the CHANGELOG entry, and proposal.md hedges the website surface with "may also apply" citing only a precedent issue number. When one bundled half of a PR (for example a security pin) has its issue pair filed and the primary change does not, the primary change's gates are still open. Verify by searching issue titles and bodies for the change name, the spec id (for example RD-FR-058), and new flag names (proposals_truncated) — label search alone is unreliable because the docs/documentation label namespace is unstandardized (#825).
