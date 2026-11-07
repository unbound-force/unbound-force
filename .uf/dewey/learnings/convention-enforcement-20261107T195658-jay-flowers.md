---
tag: convention-enforcement
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:56:58Z
identity: convention-enforcement-20261107T195658-jay-flowers
tier: draft
---

When a fix round closes a drift-test coverage gap by extending a sweep, re-enumerate the pin surfaces in the same commit: the closing commit itself can add a new surface the extended sweep still misses. Precedent: PR #957 round 3 extended the twin-pin net with TestOsvRecipeDoc_PinsMatchCanonicalJob (docs recipe ↔ ci_security.yml) and TestRemediationConstants_CarryNpm11EscapeHatch (scaffold ↔ doctor constants), while the same commit added the osv_plugin_transitive job to ci_scheduled.yml carrying npx -y npm@11.6.2, GOTOOLCHAIN=go1.27.0, and osv-scanner@v2.7.0 outside every sweep — the job's own comment claims twin status with the Makefile pin, untested. Rule: after writing or extending a drift test, grep the full diff for the pinned value string and confirm every new hit is either inside a sweep or consciously excluded with a recorded reason.
