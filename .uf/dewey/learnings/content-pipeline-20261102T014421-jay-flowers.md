---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:44:21Z
identity: content-pipeline-20261102T014421-jay-flowers
tier: draft
---

Announcement/blog issues that depend on an in-flight epic should be filed as tracking issues in the repo where the epic trigger is observable, carry an explicit trigger condition ("when #N closes"), and preserve the supporting evidence at review time rather than at write time. Triage of unbound-force/unbound-force#742 showed why: the content pack (PR-006) forbids announcing unbuilt features, so the website issue cannot be filed until the epic ships; deferring it with the evidence list (council run id, dispatch correlation id, measured ceremony reduction, LOC counts, ratification dates, gatekeeping authorization trail) captured while fresh means the eventual post rests on recorded facts instead of reconstruction, and the maintainer deferral decision stays documented in the issue body.
