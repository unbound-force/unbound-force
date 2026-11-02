---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:33:56Z
identity: issue-triage-20261102T003356-jay-flowers
tier: draft
---

Triage verdicts for feature issues that cite code anchors gain evidentiary weight when every anchor is verified against source before the verdict is issued. For unbound-force issue #801 (OpenCode v2 Phase-2 dual-emission), all cited anchors verified: the runtimeVersion struct at internal/scaffold/plugin_activation.go:19, the unpinned installOpenCode path (brew install anomalyco/tap/opencode or curl -fsSL https://opencode.ai/install | bash, no version pin) at internal/setup/setup.go:589-617, reviewPluginPackages at internal/doctor/checks.go:2186, the 9.9.9 broken-lock fixture at internal/doctor/review_plugins_test.go:408, and the embedded package.json mirror pinning @opencode-ai/plugin 1.19.0 with its provenance testdata. Anchor verification is cheap (targeted reads/greps) and separates well-researched issues from aspirational ones. Corollary: when a scope bullet has no matching done-when clause (e.g., #801's installer bullet — its acceptance criteria cover only the uf init/scaffold dimension), prefer an advisory split recommendation over NEEDS-CLARIFICATION when the work itself is unambiguous.
