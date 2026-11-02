---
tag: gate-integrity
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:35:28Z
identity: gate-integrity-20261102T033528-jay-flowers
tier: draft
---

Deduplicating security-control prose (MANDATORY GATE confirmation blocks, session-resume guards, shell-injection-prevention temp-file procedures) out of command files into an on-demand-loaded skill makes gate enforcement contingent on the skill actually loading: a skipped load step, a context-compression event dropping the reference, or a missing skill file silently removes a human-confirmation gate — the same compressed-context instruction loss the session-resume guard exists to defend against (a self-undermining extraction). Shipped gate specs also verify enforcement by structural presence (openspec/changes/harden-implement-gates: a CRITICAL RULE block within 5 lines of the AskUserQuestion instruction), which a 2-line skill reference cannot satisfy. Under Constitution Principle VI / determinism pack DE-003, prescribed gate output must be tool-rendered (precedent: format_unleash_exit, format_unleash_demo); a skill is at most interim documentation. Triage/review rule for dedup issues touching gate or security-procedure text: accept the dedup, but require the extraction target to be a tested tool (or the prose to stay inline at the point of use), an explicit gate-preservation parity acceptance criterion, and a structural-presence or seam test — "make check passes" does not exercise prompt-level gate behavior.
