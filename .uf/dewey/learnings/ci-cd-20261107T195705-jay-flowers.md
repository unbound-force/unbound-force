---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:57:05Z
identity: ci-cd-20261107T195705-jay-flowers
tier: draft
---

When a change both duplicates a pinned CI recipe into a new workflow surface and ships drift-guard tests for that recipe, the guards must enumerate every copy of the pinned values. Twin-pin tests that cover only the original surfaces (Makefile ↔ ci_security.yml, recipe-doc ↔ ci_security.yml) leave a new scheduled copy unguarded, and their failure messages actively point future lockstep bumps at the guarded files only — so the unguarded copy drifts silently (observed in PR #957: the ci_scheduled.yml osv_plugin_transitive job duplicated npm@11.6.2 / GOTOOLCHAIN=go1.27.0 / osv-scanner@v2.7.0 with no test reading that file). Review check: grep the repo for each pinned value string and confirm the drift-guard test reads every file the grep hits.
