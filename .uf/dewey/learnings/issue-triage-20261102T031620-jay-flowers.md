---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:20Z
identity: issue-triage-20261102T031620-jay-flowers
tier: draft
---

Tool-extraction issue triage (extends learning issue-triage-20261102T025933-jay-flowers): verify the procedure MECHANISM, not just prose existence — sibling refactors change how a procedure is implemented between issue creation and triage. Issue #669 described `gh api --paginate --slurp` with JQ flatten/map and `gh label list`; by triage time uf.triage-issue.md §1.1 used per-page `--paginate` with a `--jq` projection (no `--slurp` anywhere in .opencode/commands/) and the §4.2 label-existence check no longer cited `gh label list` verbatim. A drifted mechanism changes the tool contract, so the spec phase must re-derive the procedure from current prose, never from the issue's paraphrase. Second: when the extracted prose carries mandatory human-confirmation gates (e.g., "all label mutations require user confirmation" with Path A/B and duplicate-label supplementary gates), the issue remains VALID, but the spec must pin the DE-006 judgment boundary — the tool executes only the mechanical create/apply choreography after command-layer confirmation, and the gates stay in prose — because acceptance criteria phrased as "command updated to use tool" do not by themselves preserve the gates.
