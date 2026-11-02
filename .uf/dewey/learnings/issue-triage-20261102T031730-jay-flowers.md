---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:17:30Z
identity: issue-triage-20261102T031730-jay-flowers
tier: draft
---

Triage gotcha for prose→tool extraction issues that predate a sibling tool-family change (precedent: unbound-force #667 generate_pr_body, filed 2026-10-05 citing uf.finale.md lines 392-454): a later-landed change (finale-token-reduction) can split the cited procedure — template DETECTION moved into the uf_finale_pr_context gathering tool (the command no longer contains "PULL_REQUEST_TEMPLATE"; the file shrank to 399 lines so the cited range no longer exists) while the deterministic ASSEMBLY (section generation, Closes #N insertion, template-heading mapping, now at lines 232-254) remains command prose. The landed change can also overtake the issue's stated destination (issue says uf-workflow; the surrounding PR pipeline — pr_context gathering, create_pr posting — lives in uf-finale) and stale its arg sketch and token estimate (issue ~350 tokens vs epic #659 ~500). Verification protocol before VALID: (1) repo-wide grep for the proposed tool name (zero matches = unimplemented); (2) grep the command for load-bearing behaviors, not line numbers (template mapping and Closes #N emission both still present); (3) confirm both mirrored plugin copies exist (.opencode/plugins/ + internal/scaffold/assets/opencode/plugins/); (4) check the parent epic's checklist for the issue's tracked status. Verdict stays VALID when residual deterministic prose is verified; the spec phase re-baselines citations, args, destination plugin, and token deltas. Contrast #675, where the content diff exposed a misstated load-bearing behavior — here the diff found only relocation.
