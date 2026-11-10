---
tag: gatekeeping
author: jay-flowers
category: decision
created_at: 2026-11-09T21:26:42Z
identity: gatekeeping-20261109T212642-jay-flowers
tier: draft
---

Three instruction-lint IL-021 marker/body mismatches (go.md SC-006, meta-prompt.md MP-011, python.md TC-010) mix [SHOULD]/[MAY] markers with inner MUST clauses; fixing them inline changes either the rule's review severity or the force of its clause, which are Gatekeeping items under the AGENTS.md rule. PR #1001 / ADR-003 precedent: fix wording-only warnings (IL-022/IL-023) inline, defer the force-changing IL-021 fixes to a Proposed ADR that records options and a recommendation for maintainer decision, and leave severity definitions and the advisory gate unchanged until that decision lands. This keeps lint-baseline remediation from silently becoming an unauthorized gate modification.
