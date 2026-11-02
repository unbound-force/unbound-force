---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:54:35Z
identity: review-dispatch-20261102T035435-jay-flowers
tier: draft
---

Classification rules that feed tier or severity routing need per-rule regression tests: one fixture where the rule fires and one near-miss fixture where it must not. In unbound-force's review-dispatch plugin, the security_sensitive term /(^|\/)(cmd|internal)(\/|$)/ flags virtually every Go PR (all business logic lives under internal/), making the flag effectively always-true and the lightweight/standard tiers unreachable — yet the plugin suite never failed, because its only security_sensitive fixture (api/auth.ts) exercises the category route, not the path-segment regex. An over-broad term in a boolean classifier is invisible to line coverage and to outcome assertions that happen to expect true; only a near-miss fixture (e.g. internal/scaffold/assets/x.md must NOT set the flag) makes narrowing or broadening the rule detectable. When triaging or reviewing classifier changes, verify every rule term has a dedicated true-case and false-case fixture before considering the classifier tested.
