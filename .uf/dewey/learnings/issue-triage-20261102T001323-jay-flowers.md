---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:13:23Z
identity: issue-triage-20261102T001323-jay-flowers
tier: draft
---

In the unbound-force meta repo, open docs-labeled issues sharing a slash-command name are per-behavior-change companions, not duplicates. The /uf.unleash population as of 2026-11: #728 (direct review dispatch + consent gate), #739 (deterministic-tool migration), #772 (tier routing), #819 (build/test fix loop, PR #815), #840 (clarify ask gate) — five issues, five distinct behavior changes, all valid simultaneously. The authoritative duplicate check is the change's own openspec/changes/<name>/proposal Impact section, which names its companion issue numbers (e.g., the unleash-build-test-fix-loop proposal cites unbound-force/unbound-force#819 and unbound-force/website#322). When triaging documentation issues, discriminate on the underlying behavior change or PR reference, never on the command name in the title.
