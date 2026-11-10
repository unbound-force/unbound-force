---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-09T21:26:26Z
identity: instruction-lint-20261109T212626-jay-flowers
tier: draft
---

Instruction-lint remediations must sweep fenced template blocks, not just lint-flagged prose: IL-023-class vague-language rules exempt content inside code fences, so deployment templates carrying copies of a remediated phrase stay invisible to the linter. In unbound-force, the uf.init.md Step 5/6/8 template blocks (and their internal/scaffold/assets mirrors) are the injection sources for speckit/opsx command guardrails — PR #1001 fixed the phrase "language appropriate to" in four deployed files but missed the Step 6 taskstoissues guardrails template at uf.init.md:815 plus its mirror, leaving the authoritative template divergent from the text it injects (ADR-002 designates Step 6 as the source of truth). When remediating a phrase in deployed instruction text, grep the repo for every copy including fenced templates and their scaffold mirrors before declaring the remediation complete.
