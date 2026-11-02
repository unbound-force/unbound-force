---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:33:34Z
identity: documentation-gate-20261101T233334-jay-flowers
tier: draft
---

An audit checklist that hardcodes an enumeration of governance rules provides false assurance after the rules change, because nothing couples the enumeration to its source. uf.agent-brief.md's audit metric has required a "Website gate" behavioral rule since AGENTS.md renamed that gate to the current-repo Documentation gate on 2026-06-05 (cfd5b68, cross-repo requirement moved to the constitution's Cross-Repo Documentation rule); the stale entry survived ~4.5 months and now yields false positives (the audit flags the deliberately-relocated gate as missing) and false negatives (the 8-rule list omits the since-added Determinism gate and Commit scope rules), and contradicts the command's own create-mode template, which emits no Website gate. Rule enumerations consumed by audit logic need the drift discipline already applied to scaffold mirrors: derive the expected list from the authoritative source (AGENTS.md/constitution) or drift-test the enumeration against it. Hand-maintained checklists are verification-by-prose that rots silently.
