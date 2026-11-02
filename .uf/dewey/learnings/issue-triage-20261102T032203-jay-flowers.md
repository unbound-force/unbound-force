---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:22:03Z
identity: issue-triage-20261102T032203-jay-flowers
tier: draft
---

Prose→tool extraction triage: a zero-match grep for the issue's proposed tool name does not prove the work is unimplemented — the capability can ship under a different name, a different plugin, and a different tool count, so the established "confirm the proposed tool is absent from the plugin" check yields a false not-implemented signal. Precedent: unbound-force #666 proposed wait_for_ci + check_mergeability in the uf-workflow plugin (filed 2026-10-05 under epic #659); name-grep returns zero matches, yet a later phase of the same epic (finale-token-reduction, CHANGELOG "Closes: #659") shipped the capability as ONE tool — uf_finale_ci_poll in the uf-finale plugin — folding mergeability in as a pre-check, with uf.finale.md Step 7 rewritten to call it and unit/golden/smoke/Go-contract tests in place. Refined verification sequence: (1) grep the consuming command file for the cited prose (the gh commands and result fields the issue quotes) — if the prose is gone, something replaced it; (2) capability-grep the plugin source for those commands and fields (ci_poll, mergeable, "gh pr checks"), never the proposed name alone; (3) check the parent epic's closure trail — openspec/changes/*/.openspec.yaml originating_issue and CHANGELOG "Closes: #N" footers; (4) the reviewing agent's own live tool registry can corroborate a differently-named shipped tool. Decision rule, contrasting siblings #667/#668 where cited prose residue remains (VALID, re-measured per DE-004): when the cited prose is fully replaced AND the capability is present under any name, the issue is delivered — close as completed/superseded rather than routing it to work.
