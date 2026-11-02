---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:16:10Z
identity: issue-triage-20261102T031610-jay-flowers
tier: draft
---

Tool-extraction issues may overclaim "100% deterministic" for procedures whose current command prose embeds mandatory human confirmation gates. In unbound-force/unbound-force#669 (ensure_labels tool for uf.triage-issue.md), the cited label procedure carries Path A/Path B question-tool gates with fail-closed presentation (FR-006/FR-012), a supplementary duplicate-label close-semantics gate, and a command-level MUST invariant ("All labels are applied only with user confirmation"), yet the issue proposes "label creation + application in one atomic tool call" as an advantage. The extraction remains VALID — the mechanical remainder (existence check, gh label create, gh issue edit) is extractable, and DE-006 keeps judgment at the boundary by passing post-confirmation approved labels as tool parameters — but acceptance criteria need an explicit gate-preservation parity item, because tests written from the atomic-tool sketch alone would pass green while silently removing the human gate. Verification step for this issue class: when an extraction issue claims a procedure is fully deterministic, grep the cited prose for question-tool/confirmation gates before accepting the claim; an atomic mutation tool sketch plus no gate-preservation criterion is the tell.
