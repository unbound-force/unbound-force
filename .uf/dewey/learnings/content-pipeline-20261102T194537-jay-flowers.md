---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-02T19:45:37Z
identity: content-pipeline-20261102T194537-jay-flowers
tier: draft
---

Schema-alignment changes should carry a "Verified not affected" sweep in the proposal's Impact section listing every other call site and documented invocation of the re-aligned external command, and reviewers can verify it with one repo-wide grep of the drifted vocabulary. In the fix-uf-finale-ci-poll-gh-schema spec review (2026-11-02), grepping `detailsUrl` bounded every artifact layer asserting the old gh pr checks schema (plugin source, scaffold twin, test fixtures, the change's own artifacts) and returned zero matches in .opencode/commands/ and docs/ — validating the proposal's "No command-doc edits required" claim (uf.review-pr.md:125 documents name,state,description,link, all valid gh 2.91.0 fields; uf.run-pipeline-plan.md and uf.fix-ci.md carry no field lists). Pattern extends learning/issue-triage-20261101T233554's grep-verification technique from documentation-gap issues to schema-drift documentation-impact claims.
