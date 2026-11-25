---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-25T01:03:58Z
identity: determinism-20261125T010358-jay-flowers
tier: draft
as_of: "unbound-force@e74c8f8651d35f09a15c12d8e5d05a451c787ed5"
---

When tool-inventory --seams returns an empty list against a PR that modifies a plugin handler file, the change touched no tool registration surface (no tool() registration or input-schema change), so DE-007's seam-entry trigger cannot fire; the remaining determinism checks for such changes are reading the args contract for agent-pre-computed routing (DE-006) and confirming the new output-envelope fields are pinned by same-increment tests, including negative shape assertions. DE-006-positive precedent: PR #1206 (delegate-abort-recovery) modified uf_delegate_command's failure branch only — retryability routing for command-path MessageAbortedError moved INSIDE the tool (derived from durable resume state via readDelegateResumeState, with REPORT_RECOVERY_CODES as an in-tool constant set), and unleash-delegate.test.ts pinned every new envelope field (final_text, final_text_source, report_diagnostics, resume) plus its absence on non-recovery failures (expect(result).not.toHaveProperty(...)).
