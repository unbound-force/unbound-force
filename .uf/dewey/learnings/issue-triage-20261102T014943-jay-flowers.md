---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:43Z
identity: issue-triage-20261102T014943-jay-flowers
tier: draft
---

Triage verification for docs-tracking issues that follow a merged change works in three evidence layers, demonstrated by unbound-force/unbound-force#754 (docs for fix-triage-issue-dispatch-docs): (1) repo state — the issue's claim of in-change AGENTS.md/README updates (commit 0868c22d) is confirmed by the current AGENTS.md Technology Stack showing @opencode-ai/plugin 1.19.0 and the Node >= 22.22.2 engine floor; (2) tool registry — the issue's cited read_assessment_fields bounds (category/objectivity/reasoning at 128/64/4096 chars, split recommendations 16x512) match the registered tool's input contract exactly, and its DE-001 framing matches the spec's design decision D4; (3) knowledge graph — Dewey shows the spec artifacts exist (openspec/changes/fix-triage-issue-dispatch-docs/), the constitutionally required cross-repo website issue exists (unbound-force/website#312, created alongside #754), and near-neighbor issues #755/#756 are explicitly scoped follow-ups rather than duplicates. Layered cross-checks like these distinguish well-formed docs issues from aspirational or duplicate ones, and they confirm the repo's docs-gate pattern (precedent #717 → #754: change summary + checkbox deliverables + spec/originating-issue/precedent/website-issue context links).
