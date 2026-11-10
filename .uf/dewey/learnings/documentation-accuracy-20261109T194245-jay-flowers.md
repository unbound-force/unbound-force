---
tag: documentation-accuracy
author: jay-flowers
category: gotcha
created_at: 2026-11-09T19:42:45Z
identity: documentation-accuracy-20261109T194245-jay-flowers
tier: draft
---

Tool self-description strings can lag tool behavior. resolve_review_mode's embedded description (.opencode/plugins/review-dispatch/index.ts:9633) says a light check requires "no added, removed, or renamed file", but the STRUCTURAL_FILE_STATUSES constant it enforces (line 9357) also excludes "copied". Docs written from the implementation — unbound-force PR #1000's architecture.md/usage.md light-check conditions, which correctly say "adds, removes, renames, or copies no file" — can therefore be more accurate than the tool's own description string. When documenting or re-verifying a tool's conditions, cite the implementation constant or its tests as the source of record, not the tool's description prose; a docs-vs-description-string mismatch is a signal to read the code before "correcting" the docs backwards (issue #996 was fixed in exactly this direction).
