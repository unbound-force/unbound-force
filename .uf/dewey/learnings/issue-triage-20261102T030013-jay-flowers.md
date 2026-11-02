---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:00:13Z
identity: issue-triage-20261102T030013-jay-flowers
tier: draft
---

When an issue's solution prose and its acceptance criteria specify conflicting data sources, the acceptance criteria govern — and repository architecture breaks the tie objectively. unbound-force issue #672 (resolve_domain_owner) carries prose saying mappings "live as code constants inside the tool — sourced from the reviewer manifest" while its AC says the tool "reads .uf/reviewer-capabilities.yaml manifest for authoritative mappings". The manifest is an established project-owned extension point (opsx/reviewer-capabilities-init-survival; learning scaffold-ownership-20261031T145547): generated repositories may add local reviewers: entries that survive plain uf init re-runs, and Spec 005 FR-023 makes the manifest the only reviewer eligibility source. Baked-in code constants would make project-registered Divisor agents invisible to domain routing, so only the manifest-read design satisfies the extension-point contract — the prose line is a leftover from the issue's pre-revision draft (it already contains a strikethrough). Verification method for the issue's problem statement: grep .opencode/commands/ for persona names shows domain-ownership prose duplicated across uf.triage-issue.md (persona table), uf.address-feedback.md (routing rules like "Security concern raised → divisor-adversary"), and uf.review-council.md (agent/capability/scopes table duplicating the manifest verbatim — a drift risk the tool eliminates by making the manifest the single runtime source).
