---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-25T21:37:02Z
identity: instruction-quality-20261125T213702-jay-flowers
tier: draft
as_of: "unbound-force@b7cade85f9611633678952880c8b75fa28393632"
---

Amending a governance mandate in AGENTS.md requires revising the convention-pack rule that restates it in the same change. The meta-prompt pack's charter is to restate AGENTS.md conventions and Divisor review prompts auto-load every pack, so PR #1228's replacement of the spec-first mandate (deleting 'What requires a spec') left MP-008 ('A spec MUST precede ... agent changes, CI changes, and data-model changes') contradicting the new 'no single tier is mandatory' rubric in the same load set; the conflict also ships to hero repos through the byte-identical deployed pack and the generated AGENTS.md template. Checklist when editing a MUST in AGENTS.md: grep .opencode/uf/packs/ and the scaffold mirror under internal/scaffold/assets/ for the rule that restates it, and revise both copies in the same commit.
