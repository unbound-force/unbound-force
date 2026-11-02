---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:20:12Z
identity: issue-triage-20261102T022012-jay-flowers
tier: draft
---

A documentation issue describes its change's contract as it exists at filing time; a later change that supersedes the mechanism silently invalidates the issue body, so triage must verify the issue's named tools, parameters, and schemas against the current implementation — not only against the spec the issue references. Concrete case (unbound-force, issue #736, triaged 2026-11-01): filed 2026-10-15 against persist-unleash-summary, it accurately described the persist_unleash_summary tool (3 parameters, flat 7-field artifact schema); PR #745 (unleash-deterministic-tools, Oct 16) absorbed the tool into write_unleash_summary (8 parameters, richer artifact schema), and unleash-pipeline-hardening later removed the legacy name from the plugin tool map while the command-layer contract test (scaffold_test.go) pinned the old name's ABSENCE from step-10 prose. Executing the issue's checkboxes verbatim would have documented an unregistered tool name and an incomplete schema. Detection signals during triage: the referenced change directory still sits unarchived while a newer CHANGELOG entry or code comment says "mechanism of record" / "compatibility alias" / "absorbed"; contract tests assert the named tool's absence; sibling docs issues exist for the absorbing change. Correct triage outcome: NEEDS-CLARIFICATION quoting the current contract and recommending a body refresh or consolidation with the absorbing change's docs issue — never VALID-as-filed when the repo's own tests pin the named tool's absence.
