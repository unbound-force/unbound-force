---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:43:28Z
identity: documentation-pipeline-20261110T204328-jay-flowers
tier: draft
---

When a change extends a set that a doc paragraph enumerates (a security override joining toml/uuid/zod in docs/osv-transitive-scan.md), sweep every sentence in the enclosing section whose scope derives from that set — not just the enumeration line. PR #1031 (effect 4.0.0-beta.103, GHSA-chp3-j5qr-2vfr) updated the heading, advisory table, effect paragraph, npm ls verification command, and provenance enumeration, but left the same paragraph's "Rollback is a revert of the PR #957 pin commit, which restores zod 4.1.8 and toml 4.2.0" untouched and omitted rollback from the CHANGELOG entry — even though the bump's own tracking issue (#1034) lists "Rollback: revert the override commit" as a documentation deliverable and the Gatekeeping rule requires documented rollback for pin updates. Result: the canonical doc declares four overrides with provenance while documenting rollback for two, citing the wrong PR. Detection heuristic: after editing any set enumeration, read the enclosing section for each set member's name and every prior-PR reference; and diff the CHANGELOG entry's covered conditions against the tracking issue's "what needs documenting" list. No test pins doc prose, so the sweep is review's responsibility.
