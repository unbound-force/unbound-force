---
tag: code-review
author: jay-flowers
category: pattern
created_at: 2026-11-02T13:30:15Z
identity: code-review-20261102T133015-jay-flowers
tier: draft
---

When a PR's diff includes retrospective learning artifacts (.uf/dewey/learnings/*.md captured during the change's workflow), read them before deriving code-review findings: they function as a self-reported advisory registry. In the init-override-stubs-gitignore-sync council review (2026-11-02), the change's 18 learning files documented every candidate finding the reviewer re-derived from source — the ungated ensureOverrideStubs ordering after a failed .gitignore write (LOW, mitigations listed), the err-field convention split across subToolResult failure literals (LOW), the tautological exact-content stub test (LOW), the real-LookPath machine dependence in stealth tests (LOW), and the unlabelled doc-gate issue #916 — each with evidence, severity rationale, and a prescribed fix. Reading them first let the reviewer corroborate the known-advisory set and reserve findings for genuinely new issues (none found), satisfying the deduplicate-don't-re-report mandate without re-deriving each advisory. Technique: grep the diff's learning files for the changed function names and gate names, then treat each hit as a known-advisory candidate to corroborate rather than a novel finding.
