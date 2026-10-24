---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-23T20:44:09Z
identity: review-dispatch-20261023T204409-jay-flowers
tier: draft
---

Golden test `site:` labels and comments that hard-code line numbers of command files (e.g., site: "uf.address-feedback.md:314") drift silently when the referenced file gains an inserted section: the refs are non-enforced descriptive metadata, so the suite stays green while the numbers go stale (verified: a +30-line Section 1.4 insertion in .opencode/commands/uf.address-feedback.md shifted hard-coded refs in .opencode/test/golden-output-shape.test.ts and .opencode/test/dispatch-review-runs.test.ts, and one ref was already stale pre-change). Update such line refs in the same change that inserts sections, or derive them dynamically from the file at test time.
