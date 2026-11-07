---
tag: git-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:27:32Z
identity: git-tooling-20261105T162732-jay-flowers
tier: draft
---

Safe push-tool contract for agent-driven git pushes (shipped in uf-workflow git_push / git_push_force_with_lease): make `branch` required and validate ref names with a schema that rejects empty, over-255, leading '-' (option injection), whitespace, ':' (refspec separator), and the literal HEAD (would create refs/heads/HEAD); refuse with BRANCH_MISMATCH unless `symbolic-ref -q HEAD` equals refs/heads/<branch> (full-ref comparison defeats same-name tag ambiguity, detached HEAD fails closed); push an explicit `<rev-parse HEAD sha>:refs/heads/<branch>` refspec so a checkout switch between the check and the push cannot change what is pushed; for force pushes, lease on a caller-recorded expected SHA that is rev-parse verified locally (fail-closed EXPECTED_SHA_UNKNOWN, nothing pushed), defaulting to the remote-tracking ref plus --force-if-includes, or an empty-expect lease (branch must not exist) when no tracking ref exists. Parse `git push --porcelain` output only on success (exit 0 guarantees a status line per ref); regex-validate any SHA fragment taken from porcelain output before passing it to rev-parse.
