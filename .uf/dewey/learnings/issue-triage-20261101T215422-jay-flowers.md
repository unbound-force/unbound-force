---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:54:22Z
identity: issue-triage-20261101T215422-jay-flowers
tier: draft
---

Documentation-gate follow-up issues are verifiable triage targets rather than aspirational trackers when three checks pass: (1) the issue traces to a completed change's documentation-gate task — unbound-force #886 was explicitly required by task 2.4 of the reviewer-capabilities-init-survival change (one in-repo issue for the user-facing uf init behavior change, plus one cross-repo unbound-force/website issue per the constitution's Cross-Repo Documentation rule); (2) it scopes itself beyond the implementing PR's in-PR doc updates so it duplicates nothing; (3) its checkboxes map to concrete obligations created by the change's accepted trade-off. Durable pattern: scaffold disposition changes (uf init re-run flipping .uf/reviewer-capabilities.yaml from tool-owned Updated to project-owned Skipped, mirroring .uf/review-matrix.yaml) obligate the next release notes to state the disposition change and the canonical-updates-no-longer-auto-flow trade-off, and any future release adding a review persona must note the one-line manual reviewers: entry for already-seeded manifests — generated repositories learn the new semantics only from release notes once auto-flow is gone.
