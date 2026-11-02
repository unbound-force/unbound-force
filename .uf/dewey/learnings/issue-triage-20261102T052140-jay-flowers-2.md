---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:40Z
identity: issue-triage-20261102T052140-jay-flowers-2
tier: draft
---

When triaging "hardcoded version" claims against artifact-metadata writers, disambiguate format-version from producer-version semantics before ruling. The unbound-force artifact envelope carries BOTH `version` (documented "Hero version (semver)" in schemas/envelope/v1.0.0.schema.json:14) and a separate required `schema_version` field — this coexistence is what makes `const Version = "1.0.0"` in internal/artifacts/artifacts.go:34 a genuine defect (issue #556): the producer-version slot can never be correctly populated by a package constant, while a hardcoded value in a format-version-only field would be correct-by-design. Supporting evidence that settled the VALID verdict: Spec 002's quickstart example envelope shows a per-hero dynamic value ("version": "1.2.0" for gaze, "schema_version": "1.0.0"), Gaze's gaze_version is the shipped precedent, and .goreleaser.yaml:19 already injects `-X main.version={{.Version}}` into cmd/unbound-force/main.go's `version = "dev"` var — so the fix is pure wiring, not new release infrastructure. Also verified: the exported `artifacts.Version` constant has zero external consumers (no blast radius), and issue line citations drift over time (const cited at :47, now at :34) — verify substance, not line numbers.
