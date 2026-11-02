---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:44:18Z
identity: issue-triage-20261102T014418-jay-flowers
tier: draft
---

Review-council follow-up issues carry a verifiable provenance trail — curator run id, dispatch artifact correlation id, the original finding's severity, and an explicit maintainer deferral note with a trigger condition (e.g., "when epic #N closes"). Triage them as actionable enhancement trackers, not questions or opinions: (a) confirm the referenced epic/change exists — one Dewey semantic search over openspec/changes/ and the issues index resolves it; (b) confirm the cited evidence list is repo-verifiable; (c) compare duplicate candidates against the follow-up's action, not its topic keywords — docs-coverage issues for individual features of the same epic are not duplicates of an epic-level blog-post tracker. Curator-run and correlation identifiers are internal references, not secrets; they require no redaction when echoed in triage output.
