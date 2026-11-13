---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:09Z
identity: error-handling-20261107T180009-jay-flowers
tier: draft
---

When a preservation helper re-inserts user-authored data into a tool-owned file by structural anchor (byte-pattern match), the anchor-miss branch must fail loud, not return the shipped content unchanged. Seen in unbound-force PR #957: preservePluginDevDependencies re-inserts a target's devDependencies before the "\n  \"overrides\": {\n" anchor of the shipped .opencode/package.json; if a future manifest layout change moves or renames the overrides key, the anchor miss returns the shipped manifest silently stripped of the user's devDependencies — exactly the data the function exists to keep — and the first observable symptom is a broken make plugin-test, recoverable only from git. The test suite made this worse by pinning "no anchor → shipped unchanged" as intended behavior. Rule: a preservation helper's miss path should warn in the run output (or fail the sub-tool) whenever it found data to preserve but could not, so format drift surfaces at the refresh instead of downstream.
