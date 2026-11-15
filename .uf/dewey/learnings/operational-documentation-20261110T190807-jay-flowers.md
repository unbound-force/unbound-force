---
tag: operational-documentation
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:08:07Z
identity: operational-documentation-20261110T190807-jay-flowers
tier: draft
---

When a runbook recipe is the compensating control for a known tooling blind spot (uf doctor's manifest-install ignores package.json `overrides`), every new entry added to the blinded set must reach the runbook's verification recipe in the same change. The effect override change updated the runbook's advisory narrative, provenance enumeration, and rollback note but left `npm ls zod toml` unable to verify the package it remediated. When adding a member to an enumerated set that a runbook verifies, grep the runbook for every enumeration of the old set — recipe commands, expected-version sentences, override lists — and update each one.
