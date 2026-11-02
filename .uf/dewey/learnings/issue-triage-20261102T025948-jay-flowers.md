---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:59:48Z
identity: issue-triage-20261102T025948-jay-flowers
tier: draft
---

Triaging prose-to-plugin-tool extraction issues (token-reduction arc, parent #659) requires two verification passes beyond reading the issue: (1) Cited line ranges drift as sibling changes land — issue #675 cited uf.review-pr.md lines 829-993, but the fix-branch workflow had moved to lines 649-812 after the finale-token-reduction changes; always locate the named workflow by content grep (branch-name pattern, section heading), never trust issue line numbers. (2) Verify the proposed tool contract against the current command behavior it claims to preserve: #675's sketch described branching "from a PR head" with a base_branch arg defaulting to hardcoded "main", but the actual workflow branches from the PR's baseRefName (git checkout <baseRefName> then checkout -b), which is semantically required for pre-existing CI failures (they exist on base, independent of the PR). Acceptance criteria that pin only sanitization parity leave behavioral-parity details (branch point, parameter defaults, PR-derived vs hardcoded values) unresolved — these must be pinned at spec time or a refactor silently changes behavior in a git-state-mutating path.
