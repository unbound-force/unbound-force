---
tag: quick-address-feedback-diff-path-producer
author: jay-flowers
category: pattern
created_at: 2026-10-23T20:44:19Z
identity: quick-address-feedback-diff-path-producer-20261023T204419-jay-flowers
tier: draft
---

Session quick/address-feedback-diff-path-producer (2026-10-23, issue #811, follow-up to #763): anchoring a command input to its producer works best by referencing the registered tool, not by hand-rolling the producer as prose. The initial implementation of §1.4 "Fetch and Save the PR Diff" in .opencode/commands/uf.address-feedback.md prescribed mkdir/chmod/gh-redirect shell choreography; the review council (architect + guard, CRITICAL determinism) caught it as a DE-001/DE-005 violation duplicating fetch_pr_review_context. The iteration-1 fix replaced the shell block with a single intent-level fetch_pr_review_context call (returned diff_path = the immutable §2.3.4 input, failure handling mapped to the tool's structured error envelopes), added the required documentation artifacts (docs issue #814, CHANGELOG Unreleased entry), recomputed stale hard-coded line refs in the two plugin test files, and the re-review returned APPROVE WITH ADVISORIES in one iteration. User-facing command changes on the quick tier still require the documentation gate (docs issue + CHANGELOG) — the quick tier waives spec ceremony, not documentation obligations.
