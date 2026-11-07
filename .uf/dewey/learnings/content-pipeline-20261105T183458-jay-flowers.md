---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:34:58Z
identity: content-pipeline-20261105T183458-jay-flowers
tier: draft
---

Byte-identical file relocations carry pre-existing stale content into the new path: when a directory migration moves files verbatim (unbound-force PR 948's `.opencode/skill/` → `.opencode/skills/`), every line is nominally "added" in the diff yet the semantic change is only the path, so review confinement rules discourage pre-existing-content findings and factual drift survives the migration. The relocation is the cheapest refresh touchpoint — re-verify count/roster/version claims in moved files at migration time and surface drift as LOW informational findings tied to the move rather than skipping it. Verified on the unbound-force-heroes SKILL.md move, whose "5 personas" roster had drifted to 9 divisor agents by migration time.
