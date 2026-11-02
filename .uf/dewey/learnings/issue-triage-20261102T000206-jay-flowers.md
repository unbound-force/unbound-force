---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:02:06Z
identity: issue-triage-20261102T000206-jay-flowers
tier: draft
---

Triage recipe for issues that cite exact line numbers against a path that does not exist: verify the distinctive cited phrase repo-wide before concluding the issue is stale or invalid — renamed files and sibling directories routinely preserve the cited content at the cited lines. Verified case (unbound-force/unbound-force issue #824, triaged 2026-11-01): the issue cites `.opencode/commands/constitution-check.md:43,95` enumerating "the four org principles (I, II, III, IV)", but that path does not exist — the commands file is now `uf.constitution-check.md`, a thin delegation wrapper with no principle enumeration. A repo-wide grep for the cited phrase matched `.opencode/agents/constitution-check.md` at exactly lines 43 and 95 (the agent the command delegates to; the issue's "commands/" prefix was a directory misattribution, echoed from the compiled knowledge article's shorthand "constitution-check.md:43/95"). The substantive defect was real and current: the agent template hardcodes "four org principles" and "Principles checked: [count, always 4]" against the live six-principle v1.4.0 constitution, with a byte-identical scaffold twin at internal/scaffold/assets/opencode/agents/constitution-check.md that any fix must also update (drift-enforced pair). Rule: line-number precision is strong evidence the author read a real file — resolve the path (grep the distinctive phrase, check agents/ vs commands/, uf.* renames, embedded scaffold mirrors) before assessing the claim's validity.
