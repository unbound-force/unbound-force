---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:12:05Z
identity: openspec-20261104T141205-jay-flowers
tier: draft
---

Inserting a review-follow-ups phase above the final mirror phase of an OpenSpec tasks.md preserves execution order but violates MP-018's numbered-phase ordering, and appending design decisions out of numeric order (D7 before D6) does the same to design.md; both occurred in PR 949's review-dispatch-consumer-bounds change. Renumbering ripple is cheap to bound before fixing: grep the change directory for "task N.M" and "design DN" citations — in that change no task-id cross-references existed and D6/D7 were cited only from two tasks.md phase headers, so a design.md physical reorder (moving D6 above D7) had zero ripple while the tasks.md renumber touched only that file. Prefer the physical reorder when it preserves the true execution order; renumber only when numbers contradict execution order.
