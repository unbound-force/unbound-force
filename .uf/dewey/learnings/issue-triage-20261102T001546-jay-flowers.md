---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:46Z
identity: issue-triage-20261102T001546-jay-flowers
tier: draft
---

Documentation-gate issues are objectively triageable when they name their anchors. Issue #814 (document /uf.address-feedback §1.4 PR diff producer, from #811) named the exact section (§1.4 "Fetch and Save the PR Diff"), the producing tool (fetch_pr_review_context), the failure envelopes (rate_limited/api_error/timeout; prompt_size_exceeded at the 2 MiB inline cap), the renumbered sections (1.4–1.6 → 1.5–1.7), the scaffold-mirror path, and the CHANGELOG entry. Every claim was verified by direct read/grep against .opencode/commands/uf.address-feedback.md, internal/scaffold/assets/opencode/commands/uf.address-feedback.md, and CHANGELOG.md, and the documentation gap was confirmed real: docs/usage.md and docs/architecture.md mention the command but not the producer step or the tool. Pattern: file docs-gate issues for command changes as enumerated checkable deltas — new step, producing tool with error codes, renumbering map, mirror path, changelog entry — rather than a generic "update the docs" request; triage then reduces to mechanical verification before any judgment is applied.
