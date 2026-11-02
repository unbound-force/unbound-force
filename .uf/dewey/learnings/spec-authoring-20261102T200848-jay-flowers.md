---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-02T20:08:48Z
identity: spec-authoring-20261102T200848-jay-flowers
tier: draft
---

OpenSpec spec deltas in unbound-force/meta must carry capability-prefixed FR identifiers: each '### Requirement:' body opens with '**ID: XX-FR-NNN** This requirement MUST be satisfied.', and each delta opens with the MP-014 allocation header (allocated FR range, reserved range, sibling capability locations). Before approving a change, grep its specs/ directory for 'FR-', 'ID: ', and 'capability:' — the repo carries 77 compliant id instances across sibling changes, yet recent changes (unleash-pipeline-hardening, preflight-project-checks) were drafted with zero ids in their deltas. Retrofitting after archive violates MP-002's no-renumbering rule, so the omission is only cheaply fixable at review time.
