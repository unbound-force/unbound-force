---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T18:21:24Z
identity: instruction-quality-20261116T182124-jay-flowers
tier: draft
as_of: "unbound-force@9ab5fa33ff9e448615d0ec8298ee2149284f8f73"
---

When command prose tells an agent to record fields "from the tool's response", verify the named fields against the tool's actual return shape (the persisted TypeScript interface or registered tool schema) — not against intuition. Verified case: uf.triage-issue step 4.4 said "Record each created child issue number and title from the tool's response", but create_child_issue's CreateChildIssueData returns {repo, issue_number, issue_url, attempts} with no title, so the instruction directed agents to read a nonexistent field. The repair pattern: split each recorded datum by its true source — machine-assigned identifiers (issue_number, issue_url) from the tool response, human-authored text (title) from the confirmation gate's confirmed list — and cross-check the union against the consuming schema's required fields (issue-triage v1.0.0 requires {number, title, url} on child_issues_created/split_issues elements, which is exactly what the three sources supply). The same check applies to any mutation-tool prose: post_issue_comment returns {comment_id, comment_url}, so prose must not reference fields beyond those.
