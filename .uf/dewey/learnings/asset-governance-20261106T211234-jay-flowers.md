---
tag: asset-governance
author: jay-flowers
category: pattern
created_at: 2026-11-06T21:12:34Z
identity: asset-governance-20261106T211234-jay-flowers
tier: draft
---

Policy-mirror extension pattern: when a repository extends its own project-owned policy mirrors (.uf/preflight.yaml, .uf/reviewer-capabilities.yaml) beyond the embedded canonical scaffold asset, pin the divergence with compensating tests in both directions — (1) an extension lock asserting the repository copy still contains every canonical entry unchanged and in canonical order, and (2) an enrollment lock asserting every artifact the extension references exists (e.g., every divisor-* agent file in .opencode/agents/ must have a reviewer-capabilities entry, so deleting an enrollment fails the suite instead of silently dropping a persona from dispatch). Name the exempted asset set in the drift test (e.g., projectOwnedPolicyMirrors in scaffold_test.go) and document the intentional divergence in AGENTS.md so the exemption list itself stays reviewable, and keep the embedded canonical asset validated separately from the diverged repository copy (canonical/repository test split). Precedent: unbound-force PR #959 (scaffold_test.go, reviewer_capabilities_test.go, internal/schemas/ci_test.go).
