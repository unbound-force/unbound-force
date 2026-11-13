---
tag: commit-scope
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:24:17Z
identity: commit-scope-20261110T212417-jay-flowers
tier: draft
---

When a security advisory remediation (dependency pin bump) is surfaced by a PR's own CI scan and folded onto that PR's branch, thread it back into the carrying change's artifacts — a task in tasks.md plus a What Changes bullet in proposal.md carrying the advisory id, provenance, and rollback — or split it to its own quick-tier PR. A CHANGELOG entry under its own change-name records the change historically, but leaves the OpenSpec archive, the task list, and the PR body silent about a dependency change reviewers must otherwise discover by reading the diff. Precedent: PR #1006 (plugin-lock-advisories) carried the effect 4.0.0-beta.103 override (GHSA-chp3-j5qr-2vfr) as a late commit with complete provenance and rollback documentation, but no task, proposal mention, or PR-body description.
