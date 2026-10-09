---
tag: stale-command-refs
author: jay-flowers
category: pattern
created_at: 2026-09-29T21:42:00Z
identity: stale-command-refs-20260929T214200-jay-flowers
tier: draft
---

When performing textual substitution of slash-command references across a codebase, use replaceAll on the target file only after verifying that ALL occurrences in that file are in scope. For the stale-command-refs change, openspec/specs/review-council/spec.md had exactly 6 occurrences all describing current behavior, so replaceAll was safe. In cmd/unbound-force/main.go the single occurrence was in a help-text string literal. The grep verification step (task 2.1) must exclude out-of-scope directories (archived specs, changelog, migration logic, stale-ref detection tests) to avoid false positives.
