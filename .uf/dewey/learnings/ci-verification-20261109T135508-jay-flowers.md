---
tag: ci-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-09T13:55:08Z
identity: ci-verification-20261109T135508-jay-flowers
tier: draft
---

When a CI fix repairs a check that does not run on the fixing PR — because the check ships in a reusable-CI version not yet merged (pinact arrives with org-infra reusable CI v0.9.0 while ci_checks.yml pins v0.8.0) — the change record must state the actual verification surface explicitly: authoring-time API evidence, the written line's contract compliance, and a post-merge task confirming both that the check passes once it runs and that the affected job produces real output. A job that skips on a missing secret can report green with an empty report, masking a gate that never exercised the fix. Unbound-force PR #972 (tasks 4.1 and 5.1) is the worked example: the pinact fix PR could not run pinact, so its evidence was the gh api tag/SHA and blob-identity checks plus a post-merge confirmation task covering the push-to-main gaze-ai-report run.
