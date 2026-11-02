---
tag: convention-packs
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:15:27Z
identity: convention-packs-20261102T031527-jay-flowers
tier: draft
---

Unbound Force convention-pack loading has two consumer classes, and dispatch-side tooling covers only one — "build_review_prompts already loads packs" does not make command-level pack ladders redundant. Class 1, reviewer child prompts: build_review_prompts performs readdir auto-discovery of .opencode/uf/packs/*.md and injects every pack into each dispatched prompt (PR #692; openspec/changes/build-review-prompts FR-004); this already replaced uf.review-council.md's prompt-building, so the council file carries no general pack ladder — only review-voice loading for GitHub-posted prose. Class 2, command-executor composition context: uf.review-pr.md §3.9 still carries the full manual ladder (bash `test -d .opencode/uf/packs`, then read default.md, language-detect go.md/typescript.md, -custom.md, severity.md, and "pass loaded pack content to child agent prompts"), and uf.review-pr/uf.review-council/uf.triage-issue each instruct reading review-voice packs for prose the executor composes itself — a surface dispatch-side injection never reaches. Enumeration consumers also extend beyond issue #670's two named files: uf.cobalt-crush.md ("Load convention packs from .opencode/uf/packs/"), uf.agent-brief.md (dynamic pack-file listing), uf.address-feedback.md (".opencode/uf/packs/*.md"). Before implementing list_convention_packs or removing command-level ladders as redundant, survey ALL consumers repo-wide and classify each by consumer class; any ladder change must also update the byte-identical scaffold mirror under internal/scaffold/assets/opencode/commands/ (drift detection). Discovered triaging unbound-force/unbound-force#670.
