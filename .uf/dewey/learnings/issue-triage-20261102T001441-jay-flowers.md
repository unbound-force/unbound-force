---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:14:41Z
identity: issue-triage-20261102T001441-jay-flowers
tier: draft
---

Documentation-impact tracking issues (the AGENTS.md documentation gate) are triage-verifiable end to end. The pattern: (1) confirm the underlying change is user-facing, which makes the issue mandatory rather than optional; (2) check each disposition claim against repository state — the CHANGELOG entry exists under the claimed heading (e.g. ## Unreleased / ### Added, verified by listing section headings, not just finding the text anywhere in the file), the claimed doc updates appear in the named files, and "no change needed" claims are confirmed by the absence of references to the changed surface in those files; (3) confirm the implementing PR references the tracking issue and the cross-repo unbound-force/website issue exists per the constitution's Cross-Repo Documentation rule. Applied to unbound-force issue #810 (rebase-upstream-before-push): all six dispositions verified — CHANGELOG.md line 404 under Unreleased/Added, docs/usage.md workflow description plus pipeline and quick-reference rows updated, docs/architecture.md rows updated, README.md and AGENTS.md correctly unchanged (no /uf.finale step-sequence content). When every disposition verifies, the tracking issue's purpose is fulfilled and it is ready to close with the implementing PR.
