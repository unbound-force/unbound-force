---
tag: documentation-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:24:39Z
identity: documentation-drift-20261104T152439-jay-flowers
tier: draft
---

The dispatch_review_runs one-line-summary shape is documented in five command templates, not four: besides uf.review-council, uf.review-pr, uf.triage-issue, and uf.address-feedback (inline code spans that line-wrap mid-phrase, so single-phrase greps miss them), uf.unleash.md carries two more sites (spec-review dispatch ~683-687, code-review dispatch ~1621-1624), each with a byte-parity scaffold mirror under internal/scaffold/assets/opencode/commands/. When a PR changes the tool's caller-visible summary line (e.g. #946's " — error: <message> (retryable: <bool>)" suffix) and also edits the documenting paragraphs, audit every site in the same PR: grep per-clause for wrapped fragments ("dispatch_review_runs:" and "(correlation_id <id>)" separately) across .opencode/commands/ and the scaffold mirror tree, then sync the mirrors byte-identically — drift tests enforce mirror parity, but nothing enforces doc-shape currency. PR 949 (2026-11-04) shipped the suffix in the tool description, CHANGELOG, and tests while all ten command-template sites kept the success-only form; the reviewer feedback named eight of the ten.
