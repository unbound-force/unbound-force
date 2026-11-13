---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:24:19Z
identity: instruction-quality-20261110T212419-jay-flowers
tier: draft
---

When always-on instruction prose (AGENTS.md, convention packs) asserts a cross-file invariant — e.g. PR #1006's CI table row "steps kept identical to ci_security.yml" — the prose is only as trustworthy as the files it describes. Review pattern: verify the claim against the referenced file before approving, and prefer pairing such prose with a drift test that pins the invariant so the always-on context cannot silently rot after future edits. Precedent: the row's claim is enforced by TestOsvPluginTransitive_ScheduledStepsMatchCanonical, which holds the osv_plugin_transitive steps byte-identical across ci_security.yml and ci_scheduled.yml (verified line-by-line at PR head 744a39b).
