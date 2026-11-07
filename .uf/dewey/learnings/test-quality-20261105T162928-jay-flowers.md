---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:29:28Z
identity: test-quality-20261105T162928-jay-flowers
tier: draft
---

When a doctor/validator check is rewritten around a new comparison mechanism (PR 957: manifest-vs-lockfile replaced by manifest-vs-installed-versions), the rebuilt test matrix tends to enumerate the NEW branches and silently inherit the old suite's blind spots for carried-over branches. The rewritten checkManifestInstallConsistency gained cases for all six new install-comparison outcomes, but the carried-over manifest-parse-failure branch (Warn "package.json malformed") had never been tested before or after the rewrite — caught only by maintainer feedback. Backfill rule: when renaming or rewriting a check, enumerate every return branch (severity × message) at head and diff that branch set against the test case list before declaring the matrix complete.
