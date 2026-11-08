---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:16:54Z
identity: dependencies-20261106T201654-jay-flowers
tier: draft
---

An import that resolves only through transitive hoisting is a phantom dependency even when a code comment documents the arrangement: the consuming manifest never promises the package, dependency review and audit see it only as transitive, and any unrelated transitive bump can remove it or change its major version with no manifest-level review signal. Observed on unbound-force PR #959: .opencode/lib/instruction-lint.ts imports { parseDocument } from "yaml" while .opencode/package.json pins exactly two runtime dependencies (@opencode-ai/plugin, zod — enforced by TestPluginPackageInputs_ExactVersionsAndLockIntegrity); yaml@2.9.1 arrives via effect and vite hoisting, one path dev-only, so make lint-instructions and the pre-flight declared check break with MODULE_NOT_FOUND if either transitive edge drops it. Reviewer symptom: any bare import under .opencode/lib or .opencode/plugins whose package name is absent from .opencode/package.json dependencies. A pinned manifest contract that blocks declaring the dependency must be amended through its governance path, not worked around with a comment.
