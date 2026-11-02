---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:22:45Z
identity: issue-triage-20261101T222245-jay-flowers
tier: draft
---

Deferred-integration follow-up issues — where a PR ships a component (e.g., `uf gate` in PR #582) and harness wiring is deferred via design.md Non-Goals — are triageable as VALID/enhancement when they carry three traceability anchors: (1) the parent acceptance criterion being completed (#514 "Wired as validation_loop.script on write-capable harnesses"), (2) the deferral record in the delivering change's design.md (openspec/changes/uf-gate/), and (3) a machine-checkable integration contract (exact invocation `uf gate --phase <phase> --format json`; exit 0/1/2; JSON on stdout, errors on stderr). Scope items embedded as "Optional" that self-identify as needing their own spec — such as a CI coverage ratchet for internal/gate documented at 90.7% but unenforced — are split candidates: a documented coverage percentage enforces nothing (it cannot regress-fail), and an "optional" rider is an untracked deferral, indistinguishable from an abandoned one. Extracting it into its own issue gives the ratchet the spec-governed, human-approved path that gatekeeping rules require for coverage-threshold changes.
