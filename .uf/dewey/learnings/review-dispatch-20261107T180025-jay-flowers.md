---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:00:25Z
identity: review-dispatch-20261107T180025-jay-flowers
tier: draft
---

When a changeset eliminates a shared artifact (a tracked lockfile, a generated config), enumerate every surface that duplicated values derived from it and give each twin a drift pin — not only the file-based twins. Observed on unbound-force PR #957 (stop-shipping-plugin-lock): the change correctly added TestPluginTestNpmPin_MatchesSecurityScan to pin the Makefile/ci_security npm-pin twins, but the byte-identical remediation strings maintained in internal/scaffold/plugin_activation.go and internal/doctor/checks.go (deliberate cross-package mirrors, comment-synced) have no mutual equality test — doctor self-tests its own strings while the scaffold side's key clause is unpinned, so the pair can drift with no CI signal. Detection recipe: after deleting a shared artifact, grep for its derived literals (install commands, version numbers, remediation text) across packages and config files, then classify each hit as pinned (test asserts equality with its twin or its consumer's output) or comment-synced, and close the comment-synced set in the same change.
