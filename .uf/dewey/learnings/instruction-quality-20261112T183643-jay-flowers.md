---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-12T18:36:43Z
identity: instruction-quality-20261112T183643-jay-flowers
tier: draft
---

When a guardrail is rewritten to route mutations through newly extracted tools, scope its exclusivity claim to exactly the mutations those tools own. A blanket "all untrusted text reaches GitHub only through tools X and Y" contradicts neighboring steps that legitimately use other channels (fixed-enum label application via gh label create/gh issue edit, sanitized search queries via gh issue list --search), leaving an executing agent unable to satisfy both the guardrail and the step instructions. Name the owning steps in the guardrail, or explicitly enumerate the sanctioned non-tool channels (observed in /uf.triage-issue PR #1024: the reminder was narrowed to steps 4.3/4.4 correctly, but Guardrail 8's rewritten exclusivity sentence was not).
