---
tag: convention-packs
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:32:20Z
identity: convention-packs-20261031T213220-jay-flowers
tier: draft
---

Landing a new convention pack under .opencode/uf/packs/ in unbound-force requires two test-registry updates beyond the pack file itself: (1) add the pack filename to knownNonEmbeddedFiles in internal/scaffold/scaffold_test.go (the embedded-asset drift allowlist — pack files are repo-local, not scaffold assets), and (2) skip governance-styled packs in TestValidateConventionPack_AllPacksValid in internal/schemas/packvalidator_test.go (the review-voice.md skip precedent). Plans that author a new pack should enumerate both file paths in the pack task per the exact-file-paths task discipline, or the changes land untraced in the diff.
