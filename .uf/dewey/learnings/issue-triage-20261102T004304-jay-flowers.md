---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:43:04Z
identity: issue-triage-20261102T004304-jay-flowers
tier: draft
---

Centrally-filed migration phase issues: the umbrella's phase-tracking checklist is the scope-fidelity oracle, and "this repo" in the body refers to the named target sibling, not the filing repo. Verified triaging unbound-force#803 (Phase 3 — gaze, OpenCode v2 umbrella #797): the umbrella's Phase 3 checklist names exactly four parallel sibling passes (#802 dewey, #803 gaze, #804 replicator + agent-kit dual-emission, #805 website), and each phase issue should carry exactly the scope its checkbox names — #803 is pure verification with native conversion deferred to Phase 5 (umbrella D1), and only #804 carries dual-emission, so scope deltas against the umbrella checkbox are the triage signal, not the issue's existence. Body phrases like "this repo" denote the target sibling: #803's "No plugins in this repo" matches gaze's plugin-free .opencode/ surface (agents/commands/skills/uf packs only), whereas the meta repo that hosts the tracker ships .opencode/plugins/ — matching the described surface to the named repo resolves the referent without needing author clarification. Extends learning opencode-v2-migration-20261102T003021-jay-flowers (pattern verified on sibling #804).
