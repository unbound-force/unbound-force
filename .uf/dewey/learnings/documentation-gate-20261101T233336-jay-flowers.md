---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:33:36Z
identity: documentation-gate-20261101T233336-jay-flowers
tier: draft
---

Restated governance rule lists inside command files drift silently when the governance itself is renamed or restructured. unbound-force issue #825 (verified against the validated documentation-gate compiled article): AGENTS.md's "Website gate" was renamed to the current-repo Documentation gate (commit cfd5b68, Jun 5 2026) and the cross-repo requirement moved into the org constitution, yet uf.agent-brief.md's audit checklist still demanded a "Website gate" rule months later — contradicting the command's own create-mode template, which no longer inserts one, so auditing a freshly generated AGENTS.md produces a phantom missing-rule flag. Scaffold mirrors are protected by byte-parity drift tests, but prose restatements of governance inside command files carry no such guard and surface only via knowledge-base audits. When a command must restate a rule list, update every copy in the same change — the live .opencode/commands/ file AND its internal/scaffold/assets/ embedded mirror — and prefer referencing the authoritative AGENTS.md/constitution section over duplicating its contents.
