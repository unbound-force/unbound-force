---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:57:14Z
identity: issue-triage-20261101T215714-jay-flowers
tier: draft
---

When triaging a documentation-gate record issue (filed per the AGENTS.md rule requiring a documentation issue against the current repo for user-facing changes before PR merge), validity hinges on verifiable bidirectional linkage rather than prose alone. Verify four things: (1) the referenced change exists on disk (openspec/changes/<name>/ or specs/NNN-*/); (2) the claimed behavior delta is real in code — for accept-variantless-parent-model (unbound-force issue #891), the corrected "requires resolved parent model" message is live in .opencode/plugins/review-dispatch/index.ts:2323,2332 and byte-identical in the internal/scaffold/assets mirror, with regression tests in .opencode/test/finalize-error-boundaries.test.ts; (3) the CHANGELOG entry cites the change name and a Spec: path back to the change directory (CHANGELOG.md:1939); (4) both documentation-gate obligations are covered — the in-repo issue and the org-constitution website-repo issue (unbound-force/website). The exemplar closure: the change's tasks.md task 2.2 cross-cites #891 and unbound-force/website#360, so the gate record, the changelog, and the change artifacts form a closed verification loop a triager can walk in minutes.
