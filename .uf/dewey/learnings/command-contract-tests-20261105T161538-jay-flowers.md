---
tag: command-contract-tests
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:15:38Z
identity: command-contract-tests-20261105T161538-jay-flowers
tier: draft
---

When one changeset carries two prose→tool migrations, pinning one migration's consumer content does not protect the other — the pinned one manufactures false confidence for the whole PR. PR #948 migrated both the #934 skill-reference rule (pinned by TestUnleashCommand_SkillReferenceRule, added in the same PR) and the #947 push invocations (unpinned; caught only by council review across two rounds). Derive each migration's pin inventory from its own diff, not from which issue happened to carry a test task: every rewritten tool call site (arguments, stop rules, success-shape reporting), each removed raw bash form (pin its absence by matching the raw invocation form so legitimate prose mentions of the flag survive), and both canonical and embedded copies via loadCommandCopies + normalizeCommandText in internal/scaffold/command_contract_test.go. A presence-only bare-tool-name clause survives any argument drift and is not a consumer pin.
