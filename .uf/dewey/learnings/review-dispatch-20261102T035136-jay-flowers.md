---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:51:36Z
identity: review-dispatch-20261102T035136-jay-flowers
tier: draft
---

Review-dispatch carve-outs that reduce reviewer coverage or tier weight for a directory path are only safe when a compensating control enforces that directory's content invariant. For unbound-force's internal/scaffold/assets/, TestEmbeddedAssets_MatchSource (internal/scaffold/scaffold_test.go) enforces byte-parity with the canonical reviewed sources, and retaining security-baseline reviewers (adversary, guard) on a low-signal scaffold-sync category keeps the path inside security review — together these make reduced multi-agent review of the byte-copies safe. Without the parity backstop, a path-keyed exclusion becomes a blind-spot vector: any change placed under the excluded path receives reduced scrutiny by construction. Two implementation facts verified in .opencode/plugins/review-dispatch/index.ts (triage of issue #655): (1) profileDiff hard-flags every path matching /(^|\/)(cmd|internal)(\/|$)/ as security_sensitive, and security_sensitive forces the heavy tier independently of line count — so a line-count-only exclusion cannot relieve the tier for internal/scaffold/assets/ files; the carve-out must flow through classifyPath's category, not tier math elsewhere. (2) The category enum is schema-governed (.uf/schemas/review-dispatch/), so a new category is a schema-evolution change, and per the single-implementation rule for security-relevant classification, the exemption belongs in classifyPath itself so every tier-routing decision keeps tracing to one tested source of truth.
