---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:54:23Z
identity: issue-triage-20261101T215423-jay-flowers
tier: draft
---

Triage verification pattern for documentation-gate tracker issues in unbound-force/unbound-force (verified on #886, reviewer-capabilities-init-survival): a valid docs-gate tracker carries four verifiable anchors — (1) the openspec/changes/<name>/ spec directory exists with its documentation-gate task (e.g. task 2.4) marked complete, (2) the in-PR documentation surfaces named in the issue body actually mention the changed artifact in the working tree (docs/architecture.md, docs/configuration.md, CHANGELOG.md Unreleased entry), (3) the code anchor cited in the body exists (scaffold.go isToolOwned branch plus the flipped ownership-table test), and (4) the cross-repo unbound-force/website issue is referenced per the constitution's Cross-Repo Documentation rule, carrying the tutorial disposition when the change restores rather than adds capability (per learning content-pipeline-20261016T172845-jay-flowers: restored capability is not new capability, so no blog issue, but both documentation gates still apply). Checklist items that mix post-merge coherence checks, release-notes language requirements, and conditional future reminders are cohesive follow-ups to one change and do not warrant splitting.
