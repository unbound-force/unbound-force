---
tag: scaffold-ownership
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:54:48Z
identity: scaffold-ownership-20261102T035448-jay-flowers
tier: draft
---

When a scaffolded config file carries a schema `version:` field that consumer code branches on, classifying it user-owned (skip-if-present) creates silent version skew: re-init never updates it, no drift warning fires, and downstream consumers either degrade silently (zod union accepting v2 and v3 in the review-dispatch plugin) or diverge from canonical policy. The unbound-force review-matrix.yaml case (issue #649): a v2 matrix with inline model identities survives every `uf init` re-run while canonical is v3, with the skip disposition giving no staleness signal. The project's own mitigation pattern, established for reviewer-capabilities.yaml (issues #884/#886), pairs user-owned classification with a fail-closed consumer-side error naming the missing entry and the exact file to edit; a versioned structural file without that loud pairing accumulates undetected drift. Rule of thumb: user-owned classification for versioned structural config requires either drift detection at scaffold time (warn when canonical differs) or a loud consumer-side failure on version mismatch — never silent union acceptance.
