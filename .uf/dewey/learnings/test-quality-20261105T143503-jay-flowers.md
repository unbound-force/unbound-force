---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:35:03Z
identity: test-quality-20261105T143503-jay-flowers
tier: draft
---

When a contract test's doc comment claims to pin a spec requirement, enumerate the requirement's scenarios against the test's clause table at review time — a scenario-count mismatch is a mechanical check that catches omitted pins (PR 949 feedback item B: RD-FR-056 declares five scenarios, TestCommandContracts_ArtifactSourcedLessons pinned four; the "Failed finalization stores nothing" scenario had no pin while its fail-closed paragraph existed in every consumer template). Second rule for pinning consumer-template prose shared across commands: verify wording uniformity before choosing the pin mechanism. Uniform fragments (e.g., "the loaded plugin predates durable proposals") can live in the shared copy loop over loadCommandCopies, but per-template wording (the fail-closed paragraphs differ across uf.review-pr/review-council/triage-issue/address-feedback/unleash) requires per-command clause entries in the commands table — a shared-fragment pin would fail on the first template whose wording deviates.
