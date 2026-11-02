---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-02T15:17:39Z
identity: documentation-pipeline-20261102T151739-jay-flowers
tier: draft
---

When an OpenSpec change alters a user-visible tool-result contract (new error code, changed retryable semantics), the complete documentation threading spans three mutually consistent artifacts: (1) proposal.md's Impact section names every affected documentation surface with line-accurate references verified against the live file (e.g. docs/configuration.md L296-312 citing the section that exists there and stating exactly what it currently documents); (2) a numbered design decision (e.g. D10) restates the documentation delta and names the constitutional obligation; (3) tasks.md carries executable documentation tasks — a CHANGELOG entry with its content prescribed in the task, the docs-page update with the exact semantics to add, and a documentation-gate task that files BOTH tracking issues (the in-repo docs issue and the unbound-force/website issue, per the org constitution's Cross-Repo Documentation rule rather than AGENTS.md's narrower single-issue restatement) with precedent issue-pair citations and both issue numbers recorded in the PR body. Verified on the split-host-resolution-failure-classes spec review (2026-11-02): all three artifacts were consistent, the cited section and line range matched the live file exactly, and the spec avoided the known one-issue-where-two-required gotcha.
