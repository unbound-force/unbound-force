---
tag: changelog
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:26Z
identity: changelog-20261110T195026-jay-flowers
tier: draft
---

A fix-ci security-pin rider on a feature branch escapes the changelog because the feature's documentation task is scoped to the spec change only, and the rider's advisory ID typically lives nowhere in tracked files — only in the rider's commit message. Curator verification that caught this on opsx/full-lesson-proposals (an effect@4.0.0-beta.103 override for GHSA-chp3-j5qr-2vfr riding an RD-FR-058 feature branch): (1) enumerate the branch's non-feature commits (fix-ci provenance riders) and diff their subject matter against the CHANGELOG entries the branch diff adds; (2) grep the advisory ID (GHSA-*/CVE-*) across tracked files — zero hits means the security rationale is invisible in the change record; (3) check the repo's precedent for documenting CVE remediation pins in the changelog (unbound-force's stop-shipping-plugin-lock entry lists each GHSA and version bump), which sets the bar the rider must meet. The remedy is a one-clause changelog entry naming the override, the advisory, and the pin surfaces updated — much cheaper before merge than reconstructing the rationale from commit archaeology after.
