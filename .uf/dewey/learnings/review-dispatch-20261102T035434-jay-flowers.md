---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:34Z
identity: review-dispatch-20261102T035434-jay-flowers
tier: draft
---

Security-sensitivity heuristics keyed on directory path segments lose all discriminative power when the segment is a structural convention of the target ecosystem. Verified triaging unbound-force issue #656: profileDiff (.opencode/plugins/review-dispatch/index.ts lines 1227-1231) ORs /(^|\/)(cmd|internal)(\/|$)/ into security_sensitive, and the flag drives three routing decisions — diffTier returns heavy whenever it is true (lines 1195-1198), shouldAugment force-includes divisor-adversary (line 1555), and a set heavy tier cap yields a HIGH advisory (lines 2020-2031). But Go convention (and the repo's AGENTS.md) mandates internal/ for all private business logic and cmd/ for entry points, so virtually every Go PR fires the flag and pins the review at maximum weight regardless of change size (PR #653: internal/scaffold/assets/ markdown byte-copies plus one test forced a 5-agent heavy dispatch for a ~20-line meaningful change). General rule: before keying a classification or tier-escalation heuristic on a directory name, check whether the target ecosystem treats that directory as a universal convention — a convention-mandated segment is a constant, not a signal. When narrowing such a match, retain genuinely security-relevant subpaths (internal/gateway/, internal/sandbox/) and keyword patterns (secret|token|key) so the carve-out does not create a review blind spot; independent signals such as classifyPath's api|handler|middleware|routes security category continue to escalate.
