---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-06T23:58:37Z
identity: security-20261106T235837-jay-flowers
tier: draft
---

Push-target validation that rejects `/`-bearing remote names still admits `.` and `..` as local-path push destinations. Git's push-target resolution treats any token that is not a configured remote name as a URL or path, and dot-relative paths carry no slash, so a slash blacklist misses them: remote="." targets the current repository and remote=".." targets its parent directory, executing the destination's receive hooks with the pusher's privileges. Remote-name guards should additionally reject the literal values "." and ".." (and may verify the name against `git remote` enumeration when strictness outweighs convenience). Observed reviewing PR #948's follow-up: the security-20261105T163851 learning and RemoteNameSchema fixed the `/`-bearing form (remote: "/tmp/repo" now fails INVALID_INPUT), but the dot forms remain reachable on both git_push and git_push_force_with_lease.
