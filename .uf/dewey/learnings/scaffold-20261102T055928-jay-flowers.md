---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:59:28Z
identity: scaffold-20261102T055928-jay-flowers
tier: draft
---

subToolResult failure literals in unbound-force's scaffold engine split the err-field convention: writeExcludeFile (internal/scaffold/scaffold_stealth.go:287) populates subToolResult.err on write failure, while writeGitignore, ensureOverrideStub, and ensureOpencodeGitignore (internal/scaffold/scaffold.go, init-override-stubs-gitignore-sync) report only action "failed" + detail, leaving err nil. Consumers branch on the field — Run()'s stealth arm treats any sr.err as fatal ("stealth init sub-tool %q failed", scaffold.go:445-454) and printSummary renders an Error line from it (:2665) — so failures from the err-less writers are strictly less observable downstream even though their detail strings carry the same information for humans. Rule: when extending a result-struct family with failure returns, fix the err-field convention for the whole family once (or extract a shared failure-result constructor) rather than inheriting each site's pre-existing shape; an err-less failure literal looks locally fine and silently narrows downstream handling. Surfaced reviewing init-override-stubs-gitignore-sync (2026-11-02).
