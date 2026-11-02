---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:17:44Z
identity: issue-triage-20261102T031744-jay-flowers
tier: draft
---

For unbound-force plugin-tool extraction issues (Constitution Principle VI / DE-001), the acceptance criteria "Tests added" + "make plugin-test passes" are enforceable via the governance-controlled 94% statements coverage floor — but the floor only guarantees statement coverage. Happy-path-only tests satisfy it while leaving the tool's failure envelope (status error + retryable flag) and conditional behaviors (e.g. generate_pr_body's "Closes #N" emitted only when originating_issue is provided) contract-unverified. Triage should route DE-007's three seam layers into the spec's test tasks rather than relying on the floor: (1) registration surface — extend .opencode/test/registration-surface.test.ts and verify dual-copy mirror parity (.opencode/plugins/uf-workflow/ canonical + internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts embedded, drift-detected); (2) documented output shape — extend golden-output-shape.test.ts covering BOTH success and failure envelopes; (3) write→read round-trips — the calling command (uf.finale.md) consuming the returned body. These Layer 1-3 test files are the shipped DE-007 precedent from PR #745 and are the natural extension points for each new uf-workflow tool.
