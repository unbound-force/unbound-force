---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-20T17:05:58Z
identity: spec-authoring-20261120T170558-jay-flowers
tier: draft
as_of: "unbound-force@581834811fb8f5636c9a65c632b8d2d812b24fac"
---

A spec requirement that claims a checkable trigger must pin the trigger's literal syntax. Describing a signal by name only ("an explicit intake-kit provenance marker", "intake-kit testability-gated criteria") without a defined string, format, or named upstream contract leaves the acceptance criterion unobservable for every path except the ones that happen to get test fixtures, and lets two compliant implementations classify the same input differently — the exact failure the requirement's own determinism claim rules out. Pin the literal (exact label substring, exact marker string, or a cited upstream contract) in the requirement or the task's test list; whatever the tests cannot construct, the requirement cannot claim as checkable.
