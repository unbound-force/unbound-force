---
tag: contract-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:51:22Z
identity: contract-testing-20261105T225122-jay-flowers
tier: draft
---

In command-contract tests, a clause fragment spanning a line wrap inside a blockquoted passage (> prefix) can never match that passage: normalizeCommandText (strings.ToLower(strings.Join(strings.Fields(text), " "))) keeps the line-leading ">" as a standalone word token, so "per Section\n> 1.2.1" normalizes to "per section > 1.2.1" and the unwrapped fragment "per Section 1.2.1" fails against the blockquoted site. The clause can still pass by matching an identical phrase in a plain (non-blockquoted) passage elsewhere in the file, silently pinning a different site than the clause names — the suite stays green while the named site's wording is pinned only accidentally. Found auditing PR #955 iteration 3: TestCommandContracts_AddressFeedback's "resume guard confirms branch first" fragment "Confirm the working tree is on the PR head per Section 1.2.1" cannot match the SESSION-RESUME GUARD (the phrase wraps across its blockquote lines, uf.address-feedback.md:18-19) and is satisfied solely by the TodoWrite resume paragraph (lines 82-83); the guard itself stays protected only through the wrap-insensitive fragment "the re-check is unconditional on resume". Fix pattern: verify every clause fragment against normalizeCommandText output for the intended site (not the raw markdown), prefer fragments contained within a single line of the target passage, or reflow the target passage so the pinned phrase does not wrap inside a blockquote.
