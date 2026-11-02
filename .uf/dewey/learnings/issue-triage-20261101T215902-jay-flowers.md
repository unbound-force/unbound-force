---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:59:02Z
identity: issue-triage-20261101T215902-jay-flowers
tier: draft
---

Slice-deferral by issue number creates a traceability contract between specs. When a shipped change defers named items to a numbered follow-up issue (example: openspec/changes/model-matrix-delegation SC-FR-007 defers both the .opencode/agents/model-matrix-classifier.md subagent file and the meta-prompt pack to slice 2, issue #889), triage of the follow-up issue must diff its enumerated acceptance criteria against the source spec's deferral list. Items recorded as deferred-to-this-issue but absent from the issue's scope (the classifier agent file in #889) become untracked once the issue closes — a silent zero-waste violation. Remedy: before validating any slice issue, query the prior change's specs/ directory (Dewey semantic search on the issue's asset names) and flag omitted deferred items for either inclusion in scope or an explicit re-deferral with new tracking.
