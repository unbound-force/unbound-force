---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:33:40Z
identity: documentation-gate-20261101T233340-jay-flowers
tier: draft
---

Frozen governance-rule enumerations inside command-file checklists drift silently from the authoritative AGENTS.md/constitution rule set whenever gates are renamed, added, or split, and nothing catches the drift until an audit. unbound-force/unbound-force#825: uf.agent-brief.md line 386 still verifies an 8-rule list containing a "Website gate" that was renamed away on 2026-06-05 (commit cfd5b68, moving the website-issue requirement to the org constitution's Cross-Repo Documentation rule) and omits Determinism gate and Commit scope from the current 10-rule Behavioral Rules; the drift survived four months and two rule additions until the 2026-10-25 knowledge-base audit. Restated gate lists must either be derived from the authoritative source at run time or swept whenever a gate-changing change lands; mirrored scaffold copies under internal/scaffold/assets/ must be swept in the same commit; and a gate rename should file a follow-up against every instruction file that restates the rule list.
