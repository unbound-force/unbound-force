---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:01:05Z
identity: documentation-gate-20261102T020105-jay-flowers
tier: draft
---

Cross-command references that cite a step by ordinal number alone (for example "/uf.unleash Step 8" in AGENTS.md) break silently when the target command's markdown sections are renumbered: the reference keeps parsing but repoints to the wrong step. The command-prose-pilot standard (unbound-force/unbound-force, openspec/changes/command-prose-pilot) requires name-based cross-references such as "the Code Review step of /uf.unleash" — a bare number MUST NOT be the only identifier, and numbers belong at the definition site of the numbered item. When review of a scoped change discovers legacy ordinal references in files outside the change's declared scope, record the conversion as a follow-up comment on the change's documentation-gate issue instead of expanding the change's file scope; the gate issue then carries both the delivery record and the deferred fix (unbound-force/unbound-force#747).
