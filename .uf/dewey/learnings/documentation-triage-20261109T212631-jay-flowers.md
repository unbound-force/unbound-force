---
tag: documentation-triage
author: jay-flowers
category: pattern
created_at: 2026-11-09T21:26:31Z
identity: documentation-triage-20261109T212631-jay-flowers
tier: draft
---

In unbound-force, a quick-tier PR that only rewrites instruction prose (RFC 2119 markers unchanged, no new commands/workflows/capabilities) is documentation-only for content-pipeline triage: the Constitution's website-docs-issue trigger and the blog/tutorial significance thresholds do not fire, so no docs/blog/tutorial issue is required. However, the hand-curated CHANGELOG.md `## Unreleased` section still expects a per-change entry — including before/after instruction-lint baseline counts as DE-004 evidence — because a prior Unreleased entry may pin the exact lint baseline the remediation invalidates, leaving the pending-release record stale without one.
