---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:24:42Z
identity: review-dispatch-20261104T152442-jay-flowers
tier: draft
---

When a PR changes a review-dispatch tool's caller-visible output shape, the documentation update must travel to every command template that documents that shape as the executor's parsing contract — not just the tool description, CHANGELOG, and tests. Issue #946 (PR 949) added an error suffix (" — error: <message> (retryable: <bool>)") to dispatch_review_runs non-ok summary lines: the tool description, CHANGELOG, and dispatch-review-runs.test.ts were updated, but the summary-shape paragraphs in uf.triage-issue.md (fenced example), uf.review-pr.md, uf.review-council.md, and uf.address-feedback.md still showed only the success form — even though the same PR rewrote those paragraphs for a sibling summary-line semantic (the proposals-count note). The complete site set is discoverable by grepping a stable fragment of the shape ("completed, <failed> failed") across .opencode/commands/ and internal/scaffold/assets/opencode/commands/: five command pairs, including uf.unleash.md, with byte-identical mirrors enforced by drift tests, so a fix must edit both sides of each pair together. This extends the known input-list doc-drift pattern: tool-output shape docs drift the same way tool-input docs do.
