---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:05:21Z
identity: error-handling-20261107T190521-jay-flowers
tier: draft
---

Subprocess-output signature matching inherits the child tool's locale. git (like most gettext-enabled tools) localizes "fatal: not a git repository" under non-C LC_ALL/LC_MESSAGES/LANGUAGE settings, so a Contains-based benign-case classifier silently reclassifies the benign case into the failure branch on localized developer machines — a recurring false warning, not a wrong pass, when the failure branch surfaces the output. Whenever a shipped CLI classifies subprocess failures by output text, pin the classification environment (LC_ALL=C on the invocation) or classify on locale-stable signals (exit codes, `git rev-parse --is-inside-work-tree` semantics). Precedent: checkTrackedPluginLock in internal/doctor (PR #957 feedback fixes) — the English-signature match at the classification point holds on CI runners and C-locale shells only.
