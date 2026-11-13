---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-13T22:08:11Z
identity: instruction-quality-20261113T220811-jay-flowers
tier: draft
as_of: "unbound-force@4558c49ae60be6c02a7f2d23c541518ee50d7fd4"
---

Descriptive enumerations in AGENTS.md are load-bearing contracts for executing agents — verify each enumerated file against every instance of the pattern, not just the exemplar. Observed on opsx/nightly-command-evals: the new "Command evals" bullet said `.eval/fixtures/<name>/` holds a guidance.md brief, but only the two full council fixtures carry one (symlinks into .eval/lib/guidance/); the three argument-rejection fixtures legitimately omit it because their command exits before asking a question (spec EV-FR-003 makes the brief conditional on questions being asked). Fix pattern: when a documented layout varies by variant class, state the conditional explicitly ("fixtures whose command asks questions also carry a guidance.md brief") rather than listing the variant-only file in the universal enumeration — otherwise agents auditing or creating fixtures act on a false premise in both directions (false incompleteness findings, or command-mismatched briefs added to variants that never use them).
