---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T06:25:06Z
identity: determinism-20261115T062506-jay-flowers
tier: draft
---

Spec-authoring determinism pattern (verified on compaction-storm-on-a-tight-context-mode, spec-review iteration 3, zero determinism findings): an OpenSpec change that adds deterministic procedures to tool code can satisfy the determinism gate at authoring time with a three-part shape. (1) Spec deltas state the mechanical procedures as normative requirements on the tool (the executor MUST count/detect/classify), including fail-closed config-knob validation — an invalid value MUST be replaced by the documented default with a warning naming the key, the invalid value, and the default applied — so misconfiguration cannot silently disable the gate or detector. (2) tasks.md classifies every implementation and test task as judgment (authorship): code authorship is judgment work even when the resulting procedure is deterministic, so no skill/script reference is owed; the exact-command verification task is classified deterministic and names a verified repo-root Makefile reference with specific targets (check, lint-instructions), matching the DE-008 parameter contract. (3) The seam-verification task names DE-007's three layers — registration surface (config keys recognized and range-validated where knobs are registered), documented output shape (result codes, envelopes, audit fields), and write→read round-trip naming the existing consuming tool (get_progress), confirmed to exist by tool-inventory owner lookup before the tasks are authored.
