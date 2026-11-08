---
tag: documentation-consistency
author: jay-flowers
category: pattern
created_at: 2026-11-08T19:49:44Z
identity: documentation-consistency-20261108T194944-jay-flowers
tier: draft
---

When a change establishes or hardens an invocation doctrine (e.g., "run the instruction linter only through the wrapper, never node-direct"), sweep every surface the same PR touches for the superseded invocation pattern before merge: new persona/command files and their frontmatter permission blocks, AGENTS.md build-command examples, and the edited doctrine text itself. Verified on unbound-force PR 959 (feedback item 10): the wrapper-only doctrine shipped in divisor-linguist.md while the same PR added divisor-determinism.md with a node-direct Step 1 command plus a matching frontmatter bash allow, AGENTS.md gained a duplicate `make lint-instructions` entry whose "or:" alternative was the prohibited node-direct invocation, and the doctrine sentence itself was swallowed by a doubled code fence left by the edit. Two structural reasons the drift ships silently: (1) frontmatter permissions and prose form one contract — updating the prose command without the bash allow (or vice versa) leaves the doctrine half-migrated; (2) a convention guard test that pins one exemplar file (instruction-lint-preflight.test.ts reads only divisor-linguist.md) enforces the doctrine for exactly that path while sibling surfaces escape — the guard must discover its subjects the way production does (glob the directory). A cheap pre-merge check: grep the full PR diff for the superseded invocation token (e.g., `node .opencode/lib/`) — every hit outside a deliberately-exempted CLI (tool-inventory.ts, which has no wrapper) is doctrine drift.
