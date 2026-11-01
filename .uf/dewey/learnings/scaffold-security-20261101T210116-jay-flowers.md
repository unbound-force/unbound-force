---
tag: scaffold-security
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:01:16Z
identity: scaffold-security-20261101T210116-jay-flowers
tier: draft
---

Flipping a scaffolded config file from tool-owned (overwrite-on-diff) to project-owned (init-survival) is security-safe only when the config's consumer validates fail-closed at load time — the ownership flip removes the tool's auto-heal, so the safety net must move to load-time validation. model-matrix-scaffold-deployment (#889): isToolOwned("uf/model-matrix.yaml") flipped to false so team edits survive plain uf init re-runs (the #884/#886 review-matrix precedent); this is safe because the matrix loader enforces pack existence and schema validity at load time (MM-FR-001) — a preserved-but-corrupt or tampered tracked matrix cannot silently misroute delegations, it blocks loudly with MATRIX_INVALID — and --force remains the documented reset path. When reviewing any ownership flip toward user/team ownership, verify three things: (1) the consumer fails closed on invalid content rather than degrading silently, (2) the file is tracked in git so malicious edits surface in PR review, (3) an explicit reset path (--force) exists and is documented in the ownership table. (Lesson proposal from divisor-adversary, #889 review-council dispatch.)
