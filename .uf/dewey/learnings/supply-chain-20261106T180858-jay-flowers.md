---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:08:58Z
identity: supply-chain-20261106T180858-jay-flowers
tier: draft
---

When a change deletes a tracked lockfile that scheduled scanners consumed, the compensating scan must inherit the time-based trigger, not only the commit-based one. In unbound-force PR 957 (stop-shipping-plugin-lock), the throwaway-lock OSV job was added to ci_security.yml, which triggers only on push/PR to main, while the daily ci_scheduled.yml workflow — whose stated purpose is detecting newly-disclosed CVEs in already-merged dependencies — now checks out a default branch with no npm lock at all, so plugin-transitive CVEs wait for the next commit event instead of the next day. Check: when moving a security control from a scanned tracked artifact to a generated-at-scan-time artifact, map the old control's trigger conditions (especially schedule/cron) onto the new one; commit-triggered scans cannot substitute for time-triggered detection of newly-disclosed CVEs in floating dependencies.
