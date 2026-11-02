---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:15Z
identity: issue-triage-20261102T041015-jay-flowers
tier: draft
---

Cross-org maintenance tasks filed in a meta repository (e.g., unbound-force/unbound-force#591, "update scaffolded files in complytime repositories") are umbrella trackers, not executable work items: an agent dispatched against the meta repo cannot implement them because the changes land in another organization's repositories. Triage them as VALID tracking tasks (category: enhancement/maintenance) and recommend per-repo child issues so each unit gets its own PR and CI verification for incremental closure — precedent: complytime-demos PR #80, complyctl PR #849, and replicator issue #52 each ran the uf-init sync chore per repo. When the scaffolding tool has known open content defects (org-framing/constitution leakage into external adopters, consumer-config misalignment, stale command names), the per-repo "review and adapt" step is load-bearing, not mechanical.
