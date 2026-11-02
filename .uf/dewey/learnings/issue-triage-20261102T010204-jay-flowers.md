---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:02:04Z
identity: issue-triage-20261102T010204-jay-flowers
tier: draft
---

Unbound Force meta-repo triage gotcha: two documentation gates apply to user-facing changes and they target different repositories. The repo-level AGENTS.md documentation gate requires a docs issue against the current repo (unbound-force/unbound-force) before PR merge, while the org constitution v1.4.0 Cross-Repo Documentation rule additionally requires a tracking issue in the unbound-force/website repository for changes affecting user-facing behavior, hero capabilities, CLI commands, or workflows — also before the implementing PR merges. A docs-tracking issue filed in the meta repo (e.g., issue #789, the referenced deferral target for the spec-review per-finding decision gate change) satisfies the repo gate but not the constitutional website-repo requirement. When triaging or auditing docs issues in the meta repo, check whether a companion unbound-force/website issue exists or is referenced before considering documentation tracking complete.
