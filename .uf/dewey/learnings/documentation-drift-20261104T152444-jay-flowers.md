---
tag: documentation-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:24:44Z
identity: documentation-drift-20261104T152444-jay-flowers
tier: draft
---

When a PR changes a tool's caller-visible output format (e.g. dispatch_review_runs's one-line summary gaining the #946 failure suffix " — error: <message> (retryable: <bool>)"), the change must also update every decision-point example that documents that output. In this repo that means the parallel dispatch-protocol paragraphs of all four command templates (uf.review-council, uf.review-pr, uf.triage-issue, uf.address-feedback) plus their byte-identical internal/scaffold/assets/opencode/commands/ mirrors — drift tests enforce mirror parity, so a canonical-only edit fails make check. The tool description and CHANGELOG alone are not sufficient: the command templates are the procedural docs executors use to capture and parse the line, and no test enforces template-to-tool-description parity for output examples, so that drift is caught only by review. Verify a doc fix across all four templates and both copies in the same change.
