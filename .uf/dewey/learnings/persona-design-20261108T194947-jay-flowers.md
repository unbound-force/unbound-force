---
tag: persona-design
author: jay-flowers
category: pattern
created_at: 2026-11-08T19:49:47Z
identity: persona-design-20261108T194947-jay-flowers
tier: draft
---

A persona's frontmatter bash allowlist and its command prose form one invocation contract: author both against the repo's sanctioned invocation surface, copied verbatim from the doctrine-owning persona — never reconstructed from the library path. PR 959 shipped a new persona (divisor-determinism) with `node .opencode/lib/instruction-lint.ts` in both its Step 1 command and its bash allow while the same PR re-scoped divisor-linguist to the `.uf/scripts/instruction-lint` wrapper ("never by running `node` on the library directly"); the instruction-lint design decision had already pinned the wrapper as the mechanism of record for "the persona" generically, because static ESM imports crash before main() can catch them and node without type stripping exits 1 masquerading as findings — failures the wrapper's pre-checks convert to clean exit 2. A doctrine authored against one exemplar persona silently binds only that persona while its guard test pins a single hard-coded path. When adding a persona that reuses another persona's tool step, copy the whole invocation surface (command string, bash allow, denial hardening) from the doctrine owner and re-read the underlying change's design decisions — "the persona invokes the wrapper" in a design written before your persona existed means every persona, including yours.
