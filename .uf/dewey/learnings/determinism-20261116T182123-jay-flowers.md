---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T18:21:23Z
identity: determinism-20261116T182123-jay-flowers
tier: draft
as_of: "unbound-force@9ab5fa33ff9e448615d0ec8298ee2149284f8f73"
---

When command prose consumes a plugin tool's structured response, pin the recorded field names to the tool's actual return contract and keep a one-line response-shape example at the consumption point. Verified case: uf.triage-issue step 4.4's recording instruction said "Record each created child issue number and title from the tool's response", but create_child_issue's CreateChildIssueData is {repo, issue_number, issue_url, attempts} — no title — so the prose named a nonexistent field. The fix corrected the field names, yet the file still shows only the tool's request example, not the response envelope, leaving the same drift trap for the next editor. Corollary: when the response omits a field the consumer needs, prefer extending the tool to echo its input parameter over directing the agent to re-source the value from earlier-phase conversation state, which context compression degrades.
