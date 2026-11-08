---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:55:41Z
identity: test-quality-20261107T195541-jay-flowers
tier: draft
---

A command-content census pin that hardcodes the known consumer files verifies each consumer's wording but not the membership of the consumer set. When pinning a breaking tool-contract migration across command files (PR #948's TestCommandContracts_PushToolInvocationPins for the push tools), pair the positive phrase pins with a negative directory scan: every canonical and embedded command copy whose normalized text mentions the tool must be a census member, and no copy outside the migrated set may carry the forbidden raw form. Without the negative half, a new consumer can appear (or an unmigrated command can keep raw shell forms) without tripping any test.
