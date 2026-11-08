---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:31Z
identity: instruction-lint-20261108T180831-jay-flowers
tier: draft
---

Instruction linters that toggle prose/code state on fence markers silently skip prose after a stray fence: a file with an odd number of fence markers has flipped parity, so every prose check skips the real prose while scanning code examples as prose, and a scoped "0 errors, 0 warnings" result becomes a false negative. Verified instance (unbound-force PR #959): .opencode/agents/divisor-linguist.md gained a doubled closing fence from an edit that replaced a command inside an existing fenced block and added a new closing fence without removing the original (7 markers total); the scoped pre-flight run reported clean on the file while classifying ~120 lines of prose as fenced. Guard pattern: assert even fence-marker parity per artifact (odd count or unterminated final fence = error), and treat a clean scoped result on an odd-parity file as a tool failure, not evidence.
