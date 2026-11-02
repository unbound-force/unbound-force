---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:42:36Z
identity: review-dispatch-20261102T034236-jay-flowers
tier: draft
---

Refactoring shared Divisor agent boilerplate into a skill (unbound-force issue #660, divisor-common) has an unstated consumer dependency: the review-dispatch prompt builder (build_review_prompts) assembles self-contained prompts by reading agent definitions, convention packs, AGENTS.md, and the constitution from disk — content under .opencode/skills/ is outside its read set. An issue or change that slims agent definition files to reference a skill must also extend the prompt-assembly/dispatch tooling to include that skill's content in dispatched prompts (and sync scaffold-embedded mirrors of the agent files where drift-detection applies), otherwise dispatched review prompts silently lose the shared sections and the "identical output before and after" acceptance criterion fails. Triage of #660 surfaced this gap: its acceptance criteria cover file creation and behavioral equivalence but do not name the dispatch prompt-builder as an affected surface.
