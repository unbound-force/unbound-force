---
tag: code-duplication
author: jay-flowers
category: pattern
created_at: 2026-11-08T22:40:19Z
identity: code-duplication-20261108T224019-jay-flowers
tier: draft
---

When a PR extracts a shared module to de-duplicate a pattern, audit the same PR's other new files for verbatim copies of adjacent patterns before merge: PR #959 created cli-common.ts to keep CLI plumbing from drifting, yet the identical TOOL_VERSION IIFE and provenance-literal construction were copy-pasted into both consumers (tool-inventory.ts and instruction-lint.ts) in the same change. Extraction PRs should sweep the whole changeset for siblings of every pattern they centralize — especially when the duplicated block implements a spec MUST (DDP-FR-002 provenance stamping), where drift between copies silently diverges a governed contract.
