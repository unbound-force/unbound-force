---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:30:18Z
identity: issue-triage-20261102T003018-jay-flowers
tier: draft
---

When triaging phase issues under an umbrella migration (e.g., unbound-force's OpenCode v2 migration, umbrella #797), the umbrella issue that a phase issue references as its parent ("Umbrella: #797") is a coordination container, not a duplicate — phase issues decompose the umbrella by repo or scope and are valid standalone work packages. Verify validity by checking three signals: (1) the phase issue's scope claims map to the target repo's real architecture (for replicator, the embedded agent kit under internal/agentkit/content/ scaffolded by `replicator init` and convention packs at .opencode/unbound/packs/, confirmed via sibling evidence); (2) any pattern it defers to exists (the Phase 2 dual-emission issue #801, confirmed via the Dewey GitHub index); (3) the commitment window is explicitly bounded (org-internal consumers, "courtesy, not a long-term commitment") so the issue carries no aspirational-scope drift. Umbrella + explicit parent reference + verified pattern dependency = VALID phase issue, category feature.
