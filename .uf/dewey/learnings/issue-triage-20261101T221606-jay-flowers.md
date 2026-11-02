---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:16:06Z
identity: issue-triage-20261101T221606-jay-flowers
tier: draft
---

Documentation-gate issue triage can be objective rather than impressionistic: the issue's per-surface dispositions are checkable against repository context. For unbound-force #856 (docs tracker for the ported /uf.rebase-pr command), four independent surfaces corroborated the stated dispositions at triage time: (1) the AGENTS.md supplied in dispatch context shows the PR Review Commands row (`/uf.rebase-pr [N]`) and the 32-to-33 commands-count bump; (2) the indexed CHANGELOG carries the port-uf-rebase-pr entry citing "Fixes: #851, Docs: #856, unbound-force/website#340"; (3) the indexed docs/usage Quick Reference shows the /uf.rebase-pr row; (4) the duplicate search found no competing docs issue. Triage pattern: first confirm the change is user-facing so the gate applies (a new command is user-facing; internal refactoring, test-only, CI-only, and spec artifacts are exempt); second, verify BOTH coexisting gates per learning documentation-gates-20261017T222146-jay-flowers — the in-repo docs issue and the org constitution's unbound-force/website tracker; third, cross-check each stated disposition against provided repo context rather than accepting the issue's self-report; fourth, treat a fully-dispositioned tracker whose implementing PR carries the changes as VALID and close-ready, not stale.
