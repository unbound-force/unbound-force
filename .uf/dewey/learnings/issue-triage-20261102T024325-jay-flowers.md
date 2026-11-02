---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:43:25Z
identity: issue-triage-20261102T024325-jay-flowers
tier: draft
---

Compound tooling-upgrade issues that bundle generated-asset regeneration (commands, skills, schemas, scaffold asset mirrors — a content pipeline) with runtime gating features (doctor version checks, init preflight — Go code in internal/doctor and cmd/) are split candidates even when planned as a single change with a unified phase plan: the two workstreams touch disjoint file sets, verify independently, and ship independently. Assess separability by file-set overlap and independent verifiability, not by shared motivation. Observed triaging unbound-force/unbound-force#713 (openspec 1.2.0→1.14.1 upgrade): the doctor/init opencode-gating workstream (internal/doctor/checks.go, runInit in cmd/unbound-force/main.go, doctor_test.go/main_test.go) is disjoint from the asset workstream (opsx-* commands, openspec-* skills, openspec/ schema, internal/scaffold/assets/ mirrors, scaffold_test.go), and bundling both pushes a single tactical OpenSpec change past the repo's <3-story tier bound (MP-007/MP-009 escalation rule). Secondary triage signal: version-marker drift between a generator (installed CLI 1.14.1) and its generated assets (generatedBy: 1.2.0) means downstream customization anchors keyed to old stock text silently fail — treat marker drift as a latent-defect signal when triaging tooling-integration issues.
