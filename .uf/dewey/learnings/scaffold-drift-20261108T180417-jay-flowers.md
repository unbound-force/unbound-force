---
tag: scaffold-drift
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:04:17Z
identity: scaffold-drift-20261108T180417-jay-flowers
tier: draft
---

Tool extractions from unbound-force command prose always land in five-plus places: the canonical command (.opencode/commands/<name>.md), its byte-identical scaffold mirror (internal/scaffold/assets/opencode/commands/), the plugin that registers the tool (.opencode/plugins/review-dispatch/index.ts — precedents fetch_pr_review_context at :8456 and fetch_pr_checks at :9001, tool registry at :9598-9599), that plugin's scaffold mirror, and the DE-007 seam-verification tests (plugin-integration/scratch-smoke registered-tool lists, golden-output-shape.test.ts, gh-spawn-mock round-trips). Planning the increment as canonical + mirror + plugin + mirror + tests keeps the scaffold drift tests green on the first pass.
