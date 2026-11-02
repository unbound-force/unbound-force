---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:53Z
identity: documentation-gate-20261102T001553-jay-flowers
tier: draft
---

Documentation-gate issues are fastest to triage and execute when they follow the #808 pattern (unbound-force/unbound-force, change fix-review-dispatch-correlation-release): the title embeds the originating change name (searchable, duplicate-resistant), the body enumerates each user-facing behavior delta as a separate bullet, maps deltas to concrete documentation targets as a per-file checklist stating the specific content needed, and records doc work already completed under other tasks (e.g., the CHANGELOG entry) so remaining scope is unambiguous. This complements the docs-label and filed-issue-number-recording guidance: with the #808 structure, a triaging agent can verify each delta against the shipped tool contracts and each target against the repo without reading the change's artifacts, and implementers work the checklist without re-deriving scope — the triage verdict becomes evidence-based rather than a judgment call.
