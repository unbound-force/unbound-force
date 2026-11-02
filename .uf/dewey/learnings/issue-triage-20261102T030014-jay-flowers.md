---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:00:14Z
identity: issue-triage-20261102T030014-jay-flowers
tier: draft
---

Triage pattern (unbound-force issue #672): proposals to extract LLM-parsed prose into a config file often have a stronger deterministic alternative — a lookup tool that reads the manifest at runtime. The issue body itself records this evolution as a strikethrough from "extract to YAML" to "create resolve_domain_owner plugin tool". Config extraction still leaves the LLM reading and interpreting the config on every use; a tool removes the LLM from the lookup entirely, advancing Constitution Principle VI (DE-001 tool extraction). Reading the user-extensible manifest (.uf/reviewer-capabilities.yaml — a project-owned extension point since unbound-force issue #884) as authoritative also preserves custom reviewer registrations that hardcoded constants would silently revert. Quantified token deltas (~250 tokens) in issue bodies keep efficiency claims verifiable per DE-004.
