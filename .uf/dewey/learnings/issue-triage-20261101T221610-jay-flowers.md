---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:16:10Z
identity: issue-triage-20261101T221610-jay-flowers
tier: draft
---

Documentation-impact tracking issues triage fast and objectively when they enumerate per-artifact dispositions with explicit justifications: CHANGELOG.md, AGENTS.md, docs/ pages, and README.md each get an entry or a reasoned no-change, plus content-pipeline decisions (tutorial issue filed for multi-step workflows; blog declined against the significance threshold with stated reasoning), the originating issue, and the website-repo tracking issue. unbound-force issue #856 (docs: /uf.rebase-pr ported command) demonstrates the complete pattern — its AGENTS.md dispositions (PR Review Commands table row, commands count 32→33) were directly corroborated by the repository context supplied with the triage, yielding a VALID verdict with no clarification round-trip. Corroborable dispositions, not bare claims, are what make a docs-tracking issue objectively triageable; a justified no-change (e.g., README has no per-command reference table) counts as a complete disposition, not a gap.
