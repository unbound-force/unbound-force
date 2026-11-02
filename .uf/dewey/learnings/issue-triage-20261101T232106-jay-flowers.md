---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:21:06Z
identity: issue-triage-20261101T232106-jay-flowers
tier: draft
---

A docs-gate issue becomes self-verifying at triage time when its body carries four elements: the exact superseded passages quoted with file:line anchors; a per-page documentation-gate assessment table that records no-change rationales for pages assessed but not updated; the content-pipeline (blog/tutorial) skip rationale recorded at filing time; and cross-references to overlapping open issues with explicit consolidation guidance. Issue #840 in unbound-force/unbound-force (/uf.unleash clarify ask gate) demonstrates all four: the quoted exit-only sentences at docs/usage.md:152-154 and docs/architecture.md:438-440 let triage confirm the gap without re-reading the docs; the per-page table (usage.md/architecture.md update, heroes.md/AGENTS.md/README.md no-change with reasons) pre-answers the gate assessment; the recorded skip (an interaction-model refinement of an existing command meets no blog/tutorial significance threshold) prevents later duplicate content-issue filing; and the #739 consolidation note applies the disposition discipline that learning content-pipeline-20261026T172308-jay-flowers identifies as the missing ingredient when changes file competing docs issues for the same sections. Triage of an issue carrying all four elements reduces to verifying that the quoted passages exist and the source change is real.
