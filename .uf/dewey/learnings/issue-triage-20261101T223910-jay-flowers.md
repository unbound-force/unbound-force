---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:39:10Z
identity: issue-triage-20261101T223910-jay-flowers
tier: draft
---

Documentation-gate issues (per the AGENTS.md rule: file a documentation issue for user-facing changes before PR merge) can be triaged objectively without code inspection when they meet three structural tests: (a) they cite the governance gate that requires them, (b) they enumerate concrete per-file dispositions including explicit no-change rationales for skipped surfaces (e.g., README.md "no change needed" with reason), and (c) they cross-reference the originating feature issue, companion tooling issues, and the unbound-force/website tracking issue. Issue #862 (uf-fix-ci docs impact) passes all three: its claimed AGENTS.md end-state (/uf.fix-ci [N] PR Review Commands row, slash-commands count 32→33) is verifiably present in the current AGENTS.md, confirming the dispositions landed rather than remaining aspirational. Numerical counts claimed across artifacts (scaffold embedded commands 11→12, QUICKSTART 11→12, architecture docs ~46→~47) should be internally consistent for a single new scaffolded file — inconsistency across those counts is a signal to check before ruling VALID.
