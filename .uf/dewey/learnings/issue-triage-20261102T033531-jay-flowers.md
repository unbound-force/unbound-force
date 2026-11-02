---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:31Z
identity: issue-triage-20261102T033531-jay-flowers
tier: draft
---

Deduplication proposals that move command-file guard prose into a skill need two checks beyond DRY (CS-004) before the design is accepted. (1) Determinism routing: the AGENTS.md determinism gate names skills as a prohibited destination for mechanical procedural prose where an executable tool is feasible — prescribed gate output belongs in a tool-rendered template (DE-003; precedent format_unleash_exit/format_unleash_demo), fixed shell choreography like mktemp+chmod 600+--input+rm -f belongs in an executable tool with tests (DE-001; precedent uf_finale_create_pr's secure temp-file path), and only judgment policy (when to re-confirm a human gate after context compression) fits skill prose. (2) Structural-presence pinning: guard language in this repo is verification-pinned by position — openspec/changes/harden-implement-gates requires a CRITICAL RULE session-resume block within 5 lines of each AskUserQuestion gate instruction, and archive-skill-guards plus port-uf-rebase-pr carry similar "guard present" scenarios — so a "1 line replaces 12" compression can silently break those structural checks unless the spec migrates the verification form in the same increment. Verification recipe for this issue class: grep the target patterns to derive scope (do not trust the issue's file list — unbound-force#661 claimed 6 files for the session-resume guard; grep found 10, and finale carries 7 MANDATORY GATE blocks, not the claimed 6), then search openspec/changes for structural-presence requirements that pin the prose being collapsed. Verified during unbound-force/unbound-force#661 triage, 2026-11-01.
