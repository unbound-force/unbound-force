---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:13:43Z
identity: issue-triage-20261102T021343-jay-flowers
tier: draft
---

Unbound Force repos satisfy the AGENTS.md documentation gate by filing a `docs:`-prefixed issue in the current repo before PR merge for user-facing changes; the org constitution's Cross-Repo Documentation rule additionally requires a website-repo issue. These documentation-gate records follow a recognizable shape: the title/body reference the originating OpenSpec or Speckit change, the body records user-facing impact and enumerates documentation touch points (user-facing docs alignment, CHANGELOG entry under ## Unreleased, cross-repo unbound-force/website issue), with some bullets in past-tense completion form. At triage they are governance artifacts, not bug reports: classify VALID / enhancement / objective, verify the referenced change directory and `documentation` label exist, treat non-past-tense "should" bullets as the remaining actionable tracking item, and do not split them — one record per change is the point of the gate. Triage of unbound-force issue #740 (fix-tool-args-schema-mismatch, refs #732) confirmed the pattern; the underlying SDK gotcha is already captured as learning/opencode-plugin-sdk-20261015T234420-jay-flowers, and the OpenSpec change artifacts (proposal/design/tasks, all tasks [x]) corroborated every factual claim in the issue body, including the 11-tool-factory count (8 uf-workflow + 3 uf-finale).
