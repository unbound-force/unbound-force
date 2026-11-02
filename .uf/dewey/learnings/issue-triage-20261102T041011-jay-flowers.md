---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:11Z
identity: issue-triage-20261102T041011-jay-flowers
tier: draft
---

Umbrella issues that refresh scaffolded files across external adopter repositories (running `uf init` in every repo of another org) are valid tracking tasks but carry a sequencing dependency the triage should surface: `uf init` overwrites tool-owned files with the source repo's current embedded content, so open scaffold-content defects ship verbatim into every refreshed repository. For unbound-force issue #591 (complytime refresh), the open defects referenced in its thread — #526/#527/#530 (UF-org framing and constitution leaking into external adopters, content-portability CP-001–CP-003 violations), #528 (schema deployed without aligning the consumer's openspec/config.yaml), and #589 (scaffolded skills referencing pre-`uf.*` command names) — would each require the same manual correction across the ~11 remaining repos if refreshed before the fixes land, followed by a second refresh pass afterward. The issue's "review the proposed changes and adapt what is necessary" step covers interim corrections, so the task itself stays VALID; the sequencing note is an efficiency recommendation for the execution plan, not a validity blocker.
