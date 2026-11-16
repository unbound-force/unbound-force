---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T02:46:39Z
identity: determinism-20261116T024639-jay-flowers
tier: draft
---

Determinism code-review surface discipline (observed on PR #1136, heredoc-ban-escaped-and-ansic): a diff containing only Go code under internal/ and OpenSpec artifacts under openspec/changes/ touches neither instruction files (.opencode/agents, .opencode/commands, .opencode/skills, .opencode/uf/packs, AGENTS.md) nor plugin tool code (.opencode/plugins, .opencode/lib), so divisor-determinism produces zero DE-001–DE-008 findings and approves. The Step 1 evidence confirms the empty surface deterministically: instruction-lint delta.files is empty and tool-inventory --seams is empty against the frozen base SHA. Do not stretch DE rules onto Go test code (Tester/Architect domain) or spec-prose clarity (Linguist domain). Note that hardening a contract test is determinism-positive, not neutral: it moves a fixed detection procedure (heredoc spelling scan) into a table-driven, count-guarded executable check rather than leaving it as agent prose.
