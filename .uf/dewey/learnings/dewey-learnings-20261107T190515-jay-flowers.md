---
tag: dewey-learnings
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:05:15Z
identity: dewey-learnings-20261107T190515-jay-flowers
tier: draft
---

When a review-fix commit series stores its own retrospective learnings in a final chore(dewey) commit, the learnings may describe pre-fix state, making it ambiguous which prescribed remedies actually shipped. Review-time audit rule: for each stored learning that names a concrete remedy (a test name, a doc path, an issue edit), grep the fix commits for the remedy's artifact before approving — a prescription with no matching artifact is an unshipped gap; this is the review-time complement of the authoring-time rule that a learning naming a missing test or remedy must be traced to a tasks.md item or have its deferral recorded. Precedent: PR #957 council round 1 — TestOsvRecipeDoc_PinsMatchCanonicalJob, TestRemediationConstants_CarryNpm11EscapeHatch, and the second Run() in TestRun_PreservesTargetDevDependenciesWhenUpdatingManifest each landed matching their learnings, while the SC-FR-004 scenario-named acceptance test and the #964/#382 documentation-issue scope extension remained unshipped despite being prescribed by learnings in the same series.
