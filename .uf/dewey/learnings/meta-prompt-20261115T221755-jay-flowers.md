---
tag: meta-prompt
author: jay-flowers
category: pattern
created_at: 2026-11-15T22:17:55Z
identity: meta-prompt-20261115T221755-jay-flowers
tier: draft
as_of: "unbound-force@a8f038d3b312a581c6b9d9d1adf637883118d458"
evidence:
  - ".uf/artifacts/dispatch/ed55ae43-93dd-4d32-a7ee-085d3e1c8eb3.json"
  - ".opencode/commands/gaze.md"
  - ".opencode/commands/gaze-fix.md"
---

Copying advisory guidance between sibling command files silently strips the context that made it actionable. /gaze fix's "run gaze crap in the background and poll the output file" works only because Step 1's redirect commands (/tmp/gaze-fix-crap.json, /tmp/gaze-fix-quality.json) sit two lines above the note; the same sentence copied into /gaze references an output file that neither /gaze nor the gaze-reporter agent defines anywhere. When guidance must appear in multiple commands, either re-anchor each copy to that command's own concrete referents or hoist it to one shared source (a skill or reference file both executing agents load). Treat an IL-050 near-verbatim pair as a referent audit trigger, not just a dedup chore: the duplicated copies also drift — the /gaze and /gaze fix copies diverged in punctuation and heading style in the very change that added them.
