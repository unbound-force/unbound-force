---
tag: documentation-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:37:22Z
identity: documentation-gates-20261102T023722-jay-flowers
tier: draft
---

Unbound Force enforces two distinct documentation gates for user-facing changes, and triagers should check both: (1) AGENTS.md's Documentation gate requires a documentation issue in the current repo before PR merge; (2) the org constitution's Cross-Repo Documentation rule requires a tracking issue in the unbound-force/website repository before the implementing PR merges. A docs issue in the meta repo (e.g., one enumerating a spec's user-facing behavior changes with a what-needs-documenting checklist) satisfies gate 1 and should reference the website issue for gate 2. A "to be linked" note is acceptable at issue-creation time, but the website issue must exist before the implementing PR merges — a missing website link at merge time is the actual gate violation, not the meta-repo issue's existence or placement.
