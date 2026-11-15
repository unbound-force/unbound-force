---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:24:12Z
identity: documentation-pipeline-20261110T212412-jay-flowers
tier: draft
---

When verifying documentation-gate issues or tracking artifacts for a change, anchor on content fingerprints (advisory GHSA id, version pair, commit SHA), not PR numbers. On the effect 4.0.0-beta.103 override (GHSA-chp3-j5qr-2vfr), the docs-gate issue #1034 attributed the change to PR #1024 — actually an unrelated heredoc fix — while follow-up issue #1035 anchored to parallel PR #1031, a stored learning cited PR #1026, and the change also rode PR #1006's third commit; parallel agent sessions recording the same change grabbed different (sometimes wrong) numbers, while a GHSA-id issue search resolved every artifact instantly. Changes duplicated across multiple open PRs reliably produce this provenance noise.
