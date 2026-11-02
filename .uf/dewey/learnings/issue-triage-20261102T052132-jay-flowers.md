---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:32Z
identity: issue-triage-20261102T052132-jay-flowers
tier: draft
---

Issue triage (#559, unbound-force): an issue that bundles a CLI enhancement (persona-level uf init install flags) with a governance-document amendment (Constitution Principle II rewording) is a split candidate even when both parts target the same underlying concern — the code change requires the spec workflow (agent/CLI changes per AGENTS.md) while the amendment follows the constitution's own governance path (PR review, explicit semver classification where PATCH-clarification vs. MAJOR-MUST-redefinition is a reviewer decision, and mandatory hero-constitution alignment review; Dewey Principle I and Replicator Principle II carry their own Composability First principles citing the org wording). Verification corollary: issue line-number citations drift on actively-developed files (scaffold.go references were 300-500 lines stale after ~2 months) and flag inventories go stale (uf init carries 5 flags — including --stealth and --check — not the 3 cited); triage should verify the cited mechanism exists rather than the exact line number before ruling on validity, and treat a stale inventory as a correction, not an invalidation, when the substantive claim holds.
