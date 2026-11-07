---
tag: git-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:45:23Z
identity: git-tooling-20261105T144523-jay-flowers
tier: draft
---

Agent-facing git push tools should never run a bare `git push`/`git push --force-with-lease`: verify the checked-out branch with `symbolic-ref -q HEAD` (fail-closed on detached HEAD → BRANCH_MISMATCH), then push the rev-parsed HEAD SHA under an explicit refspec (`<sha>:refs/heads/<branch>`) with `--porcelain`. This binds the push to the verified commit (a checkout switch after the check cannot redirect it), makes the pushed ref explicit regardless of tracking state, and yields a structured report {remote, r  [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
