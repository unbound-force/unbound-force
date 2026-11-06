---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:01:24Z
identity: spec-authoring-20261105T170124-jay-flowers
tier: draft
---

OpenSpec change deltas in unbound-force must carry capability-prefixed FR identifiers, a requirement-numbering allocation header (allocated range plus the range reserved for future changes of the same capability), and all three delta section headings — ADDED, MODIFIED, and REMOVED, with REMOVED stating explicitly that there are no removals when empty — per meta-prompt pack MP-002/MP-013/MP-014 and AGENTS.md's "FR-NNN numbering" rule. Recent precedents: split-host-resolution-failure-classes (RD-FR-038), boun
