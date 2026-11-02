---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:57:41Z
identity: issue-triage-20261101T225741-jay-flowers
tier: draft
---

Tutorial issues covering destructive-capable git workflows can be validated at triage time by tracing the requested outline to the command contract's safety guardrails. For /uf.rebase-pr (issue #855, spec port-uf-rebase-pr), every enumerated tutorial section — argument pre-validation, PR gating checks, remotes-resolved-never-assumed, conflicts-only resolution with the no-guessing rule, git_rebase_abort clean backout, force-with-lease head-branch-only push with base-branch protection, and gh mergeable verification — mapped one-to-one to phases and guardrails in .opencode/commands/uf.rebase-pr.md, confirming the request documents shipped behavior rather than aspiration. When a tutorial issue's safety-relevant claims cannot be traced to a verifiable command contract, flag the gap before accepting the issue: tutorial prose that drifts from tool-enforced guardrails teaches unsafe operations (bare --force pushes, wrong-remote pushes) that the tool itself would have prevented.
