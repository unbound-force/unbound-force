---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:26Z
identity: issue-triage-20261102T033526-jay-flowers
tier: draft
---

A skill-versus-tool re-scope is the correct triage outcome when an issue proposes a skill for a procedure that becomes fully deterministic once judgment products exist: a skill keeps the LLM orchestrating each mechanical step, which the determinism gate (AGENTS.md, Constitution Principle VI) forbids where an executable tool is feasible. The decision test: enumerate the judgment products (composed text, human gate decisions); if supplying them as parameters makes every remaining step a fixed choreography or a lookup on structured results, the artifact is an intent-level plugin tool (determinism pack DE-005/DE-006), not a skill. unbound-force #664 exhibits the full pattern: a github-review-posting skill proposal re-scoped to post_github_review({pr_number, verdict, body, inline_comments, gate_confirmed}) — the instructed MANDATORY GATE became a required boolean parameter (structural enforcement, re-supplied after context compression), the adjacent standalone lookup-tool proposal #671 (map_verdict) was folded in as an internal constant to avoid agent-chained lookups, and the "no behavioral change" criterion moved from prose assertion to test enforcement (make plugin-test). Triage corollary: when the re-scoping governance postdates the issue filing, a re-scope comment is constitution alignment, not intent drift — assess against the current rules, not the filing-date rules.
