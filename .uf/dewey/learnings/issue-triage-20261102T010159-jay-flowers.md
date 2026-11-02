---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:01:59Z
identity: issue-triage-20261102T010159-jay-flowers
tier: draft
---

When triaging a documentation-gate issue, retrieve the originating change's tasks.md documentation-gate record as provenance before judging validity. Triage of unbound-force issue #789 (docs for the spec-review per-finding gate, change spec-review-per-finding-gate) found the change's task 3.3 recording a maintainer exemption — "no repo docs issue or website issue filed per maintainer decision" (2026-10-22) — while the docs issue itself exists, filed by the same maintainer the same day. The contradiction does not invalidate the issue: the filed issue supersedes the exemption for the repo-local documentation gate, and the stale record is bookkeeping to reconcile at resolution time; it does reveal that no website-repo issue was ever evidenced for this user-facing change (the exemption skipped the constitution's Cross-Repo Documentation gate too). Validation pattern for docs-gate issues: confirm the referenced change landed via an in-repo artifact (severity.md clarifying note, new exit-reason string present in commands/templates/tests/CHANGELOG) rather than re-reviewing the implementation; treat keyword-only duplicate candidates from fallback searches (different initiative, incidental term overlap) as non-duplicates; and note that implementation PRs often complete part of the docs sweep already (exact-label grep + affected docs page), so the issue's remaining value is the concept-phrase grep remedy — exact-label greps miss paraphrased policy text like "auto-fix LOW/MEDIUM, report HIGH/CRITICAL" (learning documentation-20261022T224957) — plus the screenshot/example audit and live-vs-point-in-time hit adjudication.
