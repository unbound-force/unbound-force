---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-08T19:24:44Z
identity: seam-verification-20261108T192444-jay-flowers
tier: draft
---

A learning that records a "Verified instance" of tool behavior against a repository is only as verified as an actual execution against that repository. On unbound-force PR #959, the shipped learning seam-verification-20261106T211243 claimed tool-inventory.ts seamReport()'s genericPlugins "resolves to an empty set for the very repository that ships it" — false at head, because the pre-existing .opencode/test/registration-surface.test.ts matches the extraction regex (four `{ id: "..." }` entries) and registration is credited. The false claim came from reading the new test file's tuple-shaped PLUGINS array and reasoning about the parse while the scratch-fixture suite stayed green: a green suite built on self-authored fixtures verifies the parser, not the repository. Recipe: before recording a verified-instance learning, execute the tool against the real artifact and capture the actual output (or cite the failing/red test that reproduces the claim); cite the command and the observed result in the learning body so a future session can re-verify. A false verified-instance learning is worse than no learning — Dewey retrieves it as prior knowledge in review sessions and it propagates the false claim into future verdicts.
