---
tag: exit-code-contracts
author: yvonne-devlin
category: gotcha
created_at: 2026-11-02T17:25:25Z
identity: exit-code-contracts-20261102T172525-yvonne-devlin
tier: draft
---

When a CLI spec defines multi-tier exit codes, allocate every environmental failure layer — argument validation, target-dir resolution, input discovery, input reads, output writing — to a tier explicitly, and mirror that allocation in the proposal's capability summary. In PR #582 (uf gate), proposal.md listed only exit codes 0 and 1 while spec FR-009 defined 0/1/2 with a dedicated permission-denied scenario; the implementation classified unreadable artifacts as exit-1 check failures, and the divergence survived two CHANGES_REQUESTED review rounds because the proposal-level summary made the three-tier contract invisible at triage time. Reconcile the summary, the requirement, and the implementation in the same PR — a governance gate whose own exit-code contract disagrees with its spec loses the authority to enforce spec-code agreement.
