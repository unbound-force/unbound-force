---
tag: scaffold-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:45:37Z
identity: scaffold-drift-20261104T154537-jay-flowers
tier: draft
---

unbound-force schema copies form a three-location graph — schemas/<type>/ registry ↔ .uf/schemas/<type>/ ↔ internal/scaffold/assets/uf/schemas/ — but only a subset of the registry ships under .uf/schemas/: the 8 uf/schemas/* entries enumerated in task42CanonicalAssets (lesson-proposal, model-matrix, review-dispatch, review-matrix v2+v3, review-verdict v2.0.0, reviewer-capabilities, sibling-repos). quality-report, hero-manifest, feedback-triage, envelope, issue-triage, review-verdict v1.0.0, and schemas/samples/*.json are registry-only. Drift guards over this graph must pin the known shipped set (derive it from task42CanonicalAssets) rather than require every registry file to have a .uf copy — a symmetric no-skip rule fails the suite. General principle: a verification guard that skips on os.IsNotExist is fail-open; absence checks belong on an enumerated known set, not a compared>0 count (found assessing PR 949 feedback on TestSchemaRegistryMatchesUFCopies).
