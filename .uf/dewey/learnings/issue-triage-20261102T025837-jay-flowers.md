---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:58:37Z
identity: issue-triage-20261102T025837-jay-flowers
tier: draft
---

When verifying documentation-issue claims against a completed OpenSpec change, cross-check tasks.md descriptions against the spec deltas and the implementation before treating a mismatch as an issue error. In the tier-gated-agent-count change, tasks.md task 2.3 records the heavy-cap security-sensitive advisory as MEDIUM severity, while spec FR-006 and the shipped review-dispatch code (advisories.push with severity: "HIGH") both specify HIGH — deliberately, per Constitution V (Security by Default), so the advisory surfaces in the verdict reason. Issue #676 correctly documented HIGH because it was verified against the spec and code. Spec FRs are the contract of record; task descriptions can carry stale values from earlier design iterations within the same change.
