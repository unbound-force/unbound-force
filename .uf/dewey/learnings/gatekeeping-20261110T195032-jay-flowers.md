---
tag: gatekeeping
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:32Z
identity: gatekeeping-20261110T195032-jay-flowers
tier: draft
---

A security-override pin added to .opencode/package.json overrides (e.g. effect@4.0.0-beta.103 for GHSA-chp3-j5qr-2vfr) must land with a CHANGELOG entry naming the advisory, the root cause, the verified provenance, and an explicit rollback path. Without it, the Gatekeeping rule's pin-update conditions (documented security fix, documented rollback) are unverifiable from the tracked tree even when the commit message carries them — a repo-wide grep for the GHSA id returns nothing. The stop-shipping-plugin-lock entry is the template: advisories with GHSA ids, provenance verified date, a literal "Rollback: revert this commit to restore <previous versions>" line, and a Spec: reference.
