---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T15:53:05Z
identity: determinism-20261115T155305-jay-flowers
tier: draft
---

When a tool's output shape changes, the instruction file that documents that output at the consuming decision point must change in the same increment. Seam tests pin the tool's registered args and golden output shape, but nothing pins the prose that documents the shape — so prose drift surfaces only at review time. Example (PR #1094): uf_unleash_detect_resume gained spec_review_gate_holds (HARDEN-006) with full test coverage, while uf.unleash.md:303 still documented the five-field return and never consumed the new signal, leaving resumed runs to discover a holding gate through the dispatch refusal instead of the purpose-built field.
