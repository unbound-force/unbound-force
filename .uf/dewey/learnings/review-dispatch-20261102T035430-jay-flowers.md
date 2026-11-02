---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:30Z
identity: review-dispatch-20261102T035430-jay-flowers
tier: draft
---

Path-based security-sensitivity heuristics must not treat language-convention directories as security signals. Go's `internal/` directory is the standard location for ALL business logic (the compiler enforces its import-visibility semantics), so a heuristic matching `/(^|\/)(cmd|internal)(\/|$)/` flags virtually every substantive Go PR as security-sensitive — forcing the heavy review tier, escalating divisor-adversary, and emitting tier-cap advisories — making the flag carry no signal. Verified in unbound-force/unbound-force issue #656: profileDiff() sets security_sensitive at .opencode/plugins/review-dispatch/index.ts:1227-1231, mirrored byte-identically at internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts:1230 (drift tests require both copies updated together). Security-relevance detection should key on content-bearing signals (secret/token/key/auth/credential/crypto path keywords, security-relevant package names such as gateway/sandbox, api/handler/middleware/routes segments) rather than import-visibility conventions.
