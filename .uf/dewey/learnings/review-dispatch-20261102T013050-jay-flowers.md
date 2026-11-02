---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:30:50Z
identity: review-dispatch-20261102T013050-jay-flowers
tier: draft
---

Dispatch data-flow documentation audits must be scoped by the target file's actual tool seams, not by the false-claim pattern corrected in sibling commands. Verified while triaging #756 (speckit-testreview audit): .opencode/commands/speckit.testreview.md contains zero dispatch_review_runs and zero consolidate_dispatch references — it drives dispatch via per-run invoke_agent plus direct finalize_review_dispatch — so audit checklist items premised on those tools are vacuous, while its real unaudited surface is the unnamed findings source in its consolidation step (§3.5) and the prepare_lesson_learning/dewey_store_learning lesson pipeline (§3.6). Structural corollary: this file's scaffold mirror is cross-repo — it is gaze-scaffolded (provenance marker "scaffolded by gaze v1.4.6"; mirror lives in gaze's internal/scaffold/assets, enforced by gaze's TestEmbeddedAssetsMatchSource), whereas unbound-force's internal/scaffold/assets/opencode/commands/ mirrors only the uf.* commands. "Scaffold mirror sync" for speckit.testreview.md therefore means syncing the gaze repo's embedded copy, not an in-repo mirror.
