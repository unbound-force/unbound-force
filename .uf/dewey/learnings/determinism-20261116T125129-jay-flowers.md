---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T12:51:29Z
identity: determinism-20261116T125129-jay-flowers
tier: draft
as_of: "unbound-force@db9c0c04b4e23c28da65ae3f5c5209e85114182a"
---

Determinism review scoping for code-to-shared-module consolidations: when deterministic logic already lives in tool code but is duplicated across consumers with divergent outputs (two plugins classifying the same gh exit code 4 oppositely — one retryable, one not), consolidating it into one shared pure-function module is NOT a prose→tool migration, so DE-007's three named seam layers do not strictly trigger. The seam equivalents to verify are: adapter-parity table tests running identical inputs through each consumer's public surface (classifyGhError in gh-error.test.ts; the review-dispatch adapter at the fetch_pr_checks public surface), plus the scaffold registration surfaces (isToolOwned/isDivisorAsset/stageReviewPluginSources) for the new lib module. The tool-inventory seam scan confirming no registered tool surface changed (empty seams list) is supporting evidence that input-schema parity is preserved. fix-feedback-reply-followups (gh-error.ts, FRR-FR-001) is the shipped precedent.
