---
tag: test-coverage
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:33:47Z
identity: test-coverage-20261105T183347-jay-flowers
tier: draft
---

When a fix's stated behavior spans a writer and a reader — e.g., a dispatch run file recording executor_effective_ms (writer) and dispatch_status preferring that field over the derived effective_ms (reader) — each half needs its own executing test. A test comment documenting the reader's behavior is not coverage: the writer-side test can pass while the reader silently regresses. In PR #950 (unbound-force), the run-file recording was tested end-to-end with the exact derived-vs-applied discrepancy (1_800_000 requested, 50 applied) and the dispatch_status reader preference got its own fixture asserting both the preference and the fallback. (From divisor-testing review of PR #950.)
