---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:01:02Z
identity: documentation-gate-20261101T230102-jay-flowers
tier: draft
---

A docs-gate issue's "no references today" grep claim is a point-in-time assertion that later changes can invalidate without the filing being wrong — verifying it against the current tree can produce a false falsification. Verified while triaging unbound-force/unbound-force #838 (2026-11-01): its 2026-10-26 body asserted grep found no run-pipeline-plan references in AGENTS.md/README.md/docs/, but docs/configuration.md:300 now names the command — the reference sits in the Delegation Model Matrix section added by the later model-matrix-delegation change (docs gate #893; post-dates the filing per CHANGELOG ordering) and describes model selection, a behavior update-run-pipeline-plan did not alter. Recipe: when a present-day grep hits a file a docs-gate issue cleared, establish chronology first (CHANGELOG entry order, docs-gate issue numbering) and read the hit's scope; a later-change hit with unrelated scope leaves both the filing claim (true then) and the no-update conclusion intact, and the new reference's accuracy belongs to the later change's docs gate — file no correction against the earlier issue. Corollary: no-update grep assertions in filed issue bodies should carry their as-of date so staleness reads as staleness.
