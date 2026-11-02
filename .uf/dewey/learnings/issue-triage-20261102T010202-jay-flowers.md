---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:02:02Z
identity: issue-triage-20261102T010202-jay-flowers
tier: draft
---

Unbound-force docs-gate issues may have their primary scope already satisfied by the implementing PR itself: PR #792 (spec-review per-finding gate, Closes #777) updated the canonical commands, severity pack, scaffold mirrors, docs/heroes.md, CHANGELOG, plugin strings, and structural pin tests in the same change, while docs-gate issue #789 stays open as the designated tracker for residual scope — screenshot/example verification plus post-approval advisories explicitly deferred to it by the retrospective learning (opsx-spec-review-per-finding-gate-20261022T225039-jay-flowers). Triage pattern: (1) grep for old-behavior strings to establish what doc work actually remains; (2) treat historical CHANGELOG entries, OpenSpec change artifacts, and completed specs as point-in-time records that legitimately describe superseded behavior — not stale user-facing docs; (3) check for binary screenshot assets before assuming the screenshot task is outstanding (this repo has zero .png files). In #789's case no stale old-flow text remained, so residual scope was advisory absorption plus closure verification, not doc rewrites.
