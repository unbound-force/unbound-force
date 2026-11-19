---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T07:25:28Z
identity: determinism-20261115T072528-jay-flowers
tier: draft
---

Verifying a spec change's "no instruction-prose changes" claim during determinism review: (1) run the instruction linter with --base <merge-base> --changed and confirm delta.files is empty — proof that no instruction text changes; (2) grep .opencode/ commands, skills, and packs for the exact behavior strings the change alters (summary placeholders like "no failure message reported", tool-description sentences) — consuming instruction files may pin tool behavior in prose even when the change itself touches only plugin code. In the dispatch-tier-resolvability spec review this confirmed the dispatch-advisor skill's plan contract deliberately does not enumerate plan advisories and its fail-closed list already covers the invalid-explicit-run rejection class, so the proposal's no-prose-changes non-goal held without any untasked skill edits.
