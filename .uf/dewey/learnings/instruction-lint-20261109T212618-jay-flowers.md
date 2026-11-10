---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-09T21:26:18Z
identity: instruction-lint-20261109T212618-jay-flowers
tier: draft
---

Clearing an instruction-lint baseline must not "fix" IL-021 marker/body mismatches by editing rule text: the mismatch (go.md SC-006 [SHOULD] with inner MUST, meta-prompt.md MP-011 [MAY] with inner MUST, python.md TC-010 [SHOULD] with inner MUST) can only be resolved by reclassifying the rule's RFC 2119 force, which is a Gatekeeping item — the constitution lists MUST/SHOULD rule classifications as governance gates an agent MUST NOT modify. The correct remediation-PR move is to fix only force-neutral wording (IL-022/IL-023 rewrites, RFC-2119-marker-preserving) and defer IL-021 fixes to a recorded maintainer decision, as ADR-003 does. An agent that "helpfully" rewrites the clause force to clear the warning is tampering with a gate.
