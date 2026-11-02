---
tag: preflight
author: jay-flowers
category: pattern
created_at: 2026-11-02T21:23:46Z
identity: preflight-20261102T212346-jay-flowers
tier: draft
---

When classifying check failures as branch-caused vs pre-existing via a merge-base worktree baseline, run the branch's check script (resolved from the primary checkout root) with the worktree as cwd, rather than re-running the worktree's copy of the script. Git-excluded script trees (e.g. stealth installs under git-excluded .uf/scripts/) do not exist in a merge-base worktree, so re-running the worktree's copy would fail every baseline and misclassify all failures. Pinning the script and varying only the code under test (via cwd) is exactly the variable causality classification asks about; expose the primary root to scripts through an env var (UF_PREFLIGHT_PROJECT_ROOT) so they can locate their config. Validated in opsx/preflight-project-checks design decision D5 with round-trip coverage (pre-existing and branch-caused classifications both verified against merge base e64a6564).
