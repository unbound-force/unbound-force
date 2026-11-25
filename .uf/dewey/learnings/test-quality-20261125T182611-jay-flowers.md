---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-25T18:26:11Z
identity: test-quality-20261125T182611-jay-flowers
tier: draft
as_of: "unbound-force@59d260231240f979d60ee4fc3d451fa5e5c39723"
---

When a sanitization pipeline documents an ordering property between its stages (e.g. "redact the whole text before cutting, so a secret that straddles the cut is never left half-visible"), the ordering itself is the contract — and it is exactly the property per-stage unit tests miss. Redaction tests use short input; truncation tests use clean input; a regression that swaps the stage order passes both while leaking the tail half of a boundary-straddling secret, because the retained fragment no longer matches the redaction regex. Lock documented stage orderings with a boundary-straddling test: place a credential so the length bound's cut lands inside its value, and assert the output contains the redaction marker and neither the raw value nor its tail half. Observed on unbound-force PR #1206 (delegate-abort-recovery): boundFinalText's redact-then-cut ordering in .opencode/lib/report-recovery.ts was documented in a comment but unpinned by any test.
