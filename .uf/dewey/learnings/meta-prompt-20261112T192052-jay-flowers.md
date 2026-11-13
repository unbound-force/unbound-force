---
tag: meta-prompt
author: jay-flowers
category: gotcha
created_at: 2026-11-12T19:20:52Z
identity: meta-prompt-20261112T192052-jay-flowers
tier: draft
---

When a DE-005 prose→tool extraction rewrites a guardrail that referenced the old prose scope, audit every externally-bound channel the instruction still sanctions before writing a tool-exclusivity claim. In PR #1024 (/uf.triage-issue), Guardrail 8's rewrite replaced a rule scoped to `gh api` calls with an unscoped "reaches GitHub only through the `post_issue_comment` and `create_child_issue` tools" claim, while step 4.2 still routed human-confirmed label text through `gh label create --description` / `gh issue edit --add-label` shell arguments and `gh issue list --search` still carried sanitized titles — an unexecutable MUST the command body disproved. Scope exclusivity claims to the operations the extracted tools own, or enumerate the sanctioned non-tool channels with the sanitization each requires.
