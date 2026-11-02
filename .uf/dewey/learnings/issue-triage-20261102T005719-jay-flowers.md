---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:19Z
identity: issue-triage-20261102T005719-jay-flowers
tier: draft
---

Spec-phase issues that plan a pinned dependency addition triage as VALID when they commit to the gatekeeping conditions at implementation time (documented rationale, verified target, rollback path, review + CI approval) and defer the exact pin version to work start rather than naming an unverified one — the gatekeeping rule binds pins at implementation, so spec-time deferral is compliant, not vague. Corroborate stated existing-pin versions against the enumerated record locations (package.json + scaffold/embedded mirror + AGENTS.md Technology Stack): stale figures propagate from unarchived spec MUST text — #714 carried "@opencode-ai/plugin 1.4.10" from the review-council-multi-model-fanout SC-FR-003 requirement after the pin moved to 1.19.0 — so an issue that explicitly corrects a stale figure against current manifests (as #799 does, citing both .opencode/package.json and internal/scaffold/assets/opencode/package.json) is evidence of verification, strengthening rather than weakening validity.
