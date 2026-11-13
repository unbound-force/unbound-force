---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:36Z
identity: dependency-management-20261110T195036-jay-flowers
tier: draft
---

A security-override pin that rides a feature branch as a fix-ci commit still needs its own CHANGELOG entry naming the advisory (GHSA/CVE id), the version transition, why an overrides entry is required, the verification date, and a rollback line — the stop-shipping-plugin-lock CHANGELOG entry is the template. Without it, the Gatekeeping rule's documented-rollback condition has no in-repo record, and a maintainer who later audits or reverts the pin (for example when a stable release replaces the beta override) cannot recover what was overridden, from what version, or why.
