---
tag: golden-tests
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:40:42Z
identity: golden-tests-20261101T234042-jay-flowers
tier: draft
---

Census-drift verification in the golden-output-shape suite must count executable test content, not prose banners, and must check for sites the registry itself never counted. Triaging issue #823 (2026-11-01) showed the drift runs one site deeper than reported: the review-dispatch describe title ("15 documented sites", .opencode/test/golden-output-shape.test.ts:1112) matches the block's actual 15 site tests — 7 plan_review_dispatch, 3 consolidate_dispatch, 2 finalize_review_dispatch, acquire_sibling_evidence, prepare_lesson_learning, plus capture_review_diff [uf.unleash.md:620] — so the "14" section banner (line 1036) and the file-header attribution (line 60, whose tool enumeration omits capture_review_diff) are the stale labels. capture_review_diff is also absent from metrics.md's 47-site fixture inventory (uf.unleash.md is counted at 18 sites without it), so the mechanically recounted total is likely 49, not the golden header's 48: the fix cannot land on either 47 or 48 and must decide capture_review_diff's census status under the task-2.1 inclusion rule, refreshing the per-command counts in the same commit.
