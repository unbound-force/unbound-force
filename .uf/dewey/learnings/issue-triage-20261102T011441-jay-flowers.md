---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:41Z
identity: issue-triage-20261102T011441-jay-flowers
tier: draft
---

Pack-related docs-gate records in unbound-force/unbound-force split on a verifiable boundary the 2026-10-14 gotcha (documentation-20261014T123134) left implicit: adding a RULE to the existing repo-local determinism pack (explicitly marked "repo-local governance pack (not a scaffold asset)", version-bumped rather than newly created, no constitution amendment, no README-referenced surface) is classified internal/test-architecture and website-exempt — the classification #781 records for DE-007 and that #762 set for DE-005/DE-006. By contrast, determinism-codification required a website issue and a README principle-count grep because it added a NEW pack, a NEW skill, and Principle VI to the constitution. Triage recipe for exemption-record issues like #781: (1) confirm the referenced change exists (GitHub issue number and the openspec/changes/<name>/ directory on disk), (2) confirm no newly-created pack, constitution principle, or README-referenced surface is touched, and (3) confirm the exemption rationale is carried in the change's metrics.md per DE-004. Separately, duplicate candidates sharing docs/pack/skill keywords but referencing a different underlying change number or branch (e.g., #791 delegation-skill-enforcement vs #781 seam-verification) are topical keyword matches, not duplicates — compare the referenced change before classifying.
