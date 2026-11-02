---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:33:45Z
identity: issue-triage-20261102T013345-jay-flowers
tier: draft
---

Documentation-gate triage in unbound-force: one shipped increment (e.g., fix-review-dispatch-recovery / PR #776) intentionally produces a trio of related doc issues — an in-repo umbrella tracking issue (#766, tool-reference coverage for the full change set), a narrower in-repo tutorial issue (#767, workflow narrative), and a cross-repo website issue (unbound-force/website#313, per the constitution's Cross-Repo Documentation rule). Title-keyword duplicate searches surface the siblings as duplicate candidates, but they are complementary deliverables, not duplicates. Verify objectively before ruling: confirm the named tools against the live MCP tool registry, grep docs/ for the tool names to confirm the documentation gap, and confirm the openspec/changes/<name>/ directory exists. Recommend cross-linking with explicit scope ownership (reference coverage vs tutorial vs website) rather than closing as duplicate.
