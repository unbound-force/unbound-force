---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:15:48Z
identity: test-quality-20261110T231548-jay-flowers
tier: draft
---

When a compensating test joins two surfaces that must stay in lockstep (e.g., a manifest's overrides block and its provenance records), iterating one collection and asserting each entry is declared in the other enforces only one direction. An entry added to the checked-against surface — with the test's own hardcoded expectation map updated in the same commit — passes every test while violating the contract the join exists to enforce. Assert set equality in both directions: after the per-entry loop, compare collection sizes (or key sets) so an entry on either surface without its counterpart fails. Precedent: TestPluginPackageInputs_ApprovedMaintenanceProvenance joined provenance records to the scaffolded manifest overrides one-way, leaving a manifest override without a provenance record unenforced even though the Gatekeeping rule requires provenance for every pin update.
