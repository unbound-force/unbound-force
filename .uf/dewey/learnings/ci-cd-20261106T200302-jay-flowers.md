---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:03:02Z
identity: ci-cd-20261106T200302-jay-flowers
tier: draft
---

Deleting a tracked lockfile silently narrows every reader of that lockfile, not just the primary scanner. Before removing one, enumerate all consumers across trigger cadences — PR-time scans, push-time scans, and scheduled/daily scans whose stated purpose is post-merge disclosure detection — and mirror any compensating scan into each cadence. A compensation job added only to the PR workflow leaves the scheduled scan blind to exactly the class of CVEs it exists to catch: transitive, post-merge disclosures that no direct-pin scan can see (override-pinned transitives like toml/uuid and nested copies forced by overrides do not appear in package.json dependencies). Seen in unbound-force PR #957: the throwaway-lock OSV job covers push/pull_request only, while ci_scheduled.yml's daily scan lost npm transitive coverage when the tracked .opencode/package-lock.json was removed.
