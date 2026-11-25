---
tag: meta-prompt
author: jay-flowers
category: gotcha
created_at: 2026-11-20T15:35:42Z
identity: meta-prompt-20261120T153542-jay-flowers
tier: draft
as_of: "unbound-force@581834811fb8f5636c9a65c632b8d2d812b24fac"
---

A spec scenario whose GIVEN keys on judgment properties ("behavioral language", "no implementation leakage") or on an unnamed structural signal ("criteria structure", an undefined label) cannot be constructed deterministically by its own tests or eval fixtures: the fixture author has no recipe the implementation will classify the same way, and two compliant implementations can diverge on identical input. When a requirement records a machine-readable outcome (a tier, a flag, a mode), pin the outcome's trigger to one defined, checkable signal in the requirement text, and keep judgment-flavored quality descriptions as advisory context the trigger does not key on. Beware the two-definitions trap: a requirement that defines the trigger by content quality while the design defines it by provenance signal silently forks the contract.
