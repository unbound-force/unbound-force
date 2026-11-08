---
tag: contract-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T21:49:16Z
identity: contract-testing-20261105T214916-jay-flowers
tier: draft
---

When a command-file fix establishes a tool-call invariant across multiple instruction sites — the main-flow section, the session-resume guard, and companion tracking/TodoWrite sections — pin each corrected site's conditional wording with its own phrase-presence contract clause, in addition to the invariant statement and any ordering pin. The ordering pin's anchor must skip legitimately-early mentions (a resume guard references the governed tool before the gated section while ordering the actual call after it), which structurally excludes those early sites from the pin's protection; only per-site clauses lock them. Precedent: PR #955's uf.address-feedback contract test pinned Section 1.2.1's content, the 'never call get_progress or checkpoint from another branch' rule, and a 1.2.1-before-first-CHECKPOINT ordering pin, but added no clause for the corrected resume-guard wording ('the re-check is unconditional on resume') or the TodoWrite resume text — the two sites whose unfixed state re-opened issue #954 during the PR's own feedback round. A future revert of either resume text passes every pinned clause because the invariant phrase survives elsewhere in the file.
