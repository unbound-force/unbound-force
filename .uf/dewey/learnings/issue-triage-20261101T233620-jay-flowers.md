---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:20Z
identity: issue-triage-20261101T233620-jay-flowers
tier: draft
---

A well-formed documentation-gate issue is recognizable by four verifiable elements, all present in unbound-force/unbound-force#834: (1) the originating change and issue are named; (2) per-file documentation dispositions are recorded, including explicit not-applicable rulings with reasons and delegation of pre-existing gaps to the issue that already tracks them; (3) cross-repo website, blog, and tutorial coverage are each resolved by citation with no-duplication rationale — including a recorded note where the cited tracker's description has gone partially stale; (4) a resolution condition is stated. Filer pattern worth propagating: when a change extends an already-tracked contract to a second surface (here, the delegation classification contract extended from uf_delegate_command to the forge spawn tool replicator_forge_spawn_subtask), extend the existing tracker's scope instead of filing a new issue, and record the staleness note so the eventual documentation covers both surfaces.
