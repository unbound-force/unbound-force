---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:26Z
identity: review-dispatch-20261102T035426-jay-flowers
tier: draft
---

Go's `internal/` directory is a compiler-enforced import-visibility (encapsulation) boundary, not a security boundary — conflating the two breaks security-sensitivity heuristics. In unbound-force (and sibling Go hero repos gaze/dewey/replicator), AGENTS.md mandates ALL business logic under `internal/`, so the review-dispatch `profileDiff()` regex `(^|\/)(cmd|internal)(\/|$)` (`.opencode/plugins/review-dispatch/index.ts`) makes `security_sensitive` true for virtually every Go source change. Downstream effects: `diffTier()` forces the heavy tier (`securitySensitive || lines > 300`), `shouldAugment()` force-augments divisor-adversary, and a HIGH tier-cap advisory fires constantly — destroying the flag's discriminating signal and wasting heavy-tier review budget on trivial diffs (observed: PR #653, byte-copy markdown under `internal/scaffold/assets/`). The `cmd` path segment has the same false-positive problem (every CLI entrypoint is "security-sensitive"). Path-based security heuristics must match security-relevant semantics — secrets/tokens/keys keywords, auth/credential/crypto paths, or genuinely security-critical packages like `internal/gateway/` (LLM proxy, tokens) and `internal/sandbox/` (container isolation) — rather than language visibility conventions. When narrowing such a classifier, keep the precise signals (`/(secret|token|key)/`, `classifyPath`'s security category) and have the change reviewed so it does not under-trigger on genuinely sensitive paths.
