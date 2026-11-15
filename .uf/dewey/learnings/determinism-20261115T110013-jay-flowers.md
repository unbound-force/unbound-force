---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T11:00:13Z
identity: determinism-20261115T110013-jay-flowers
tier: draft
---

tool-inventory --seams classifies test files into DE-007 seam layers by FILENAME regexes (/registration-surface/, /golden|output-shape/, /round-?trip/), with a round-trip fallback on test titles. Substantive seam coverage placed in other files (e.g., unleash-delegate.test.ts's "uf_delegate_command seam verification" and "compaction-storm surfacing" describes, which pin exact envelope shapes and a write→read round-trip through get_progress) reads as layer "other", so the evidence of record reports the layers as `missing` for a tool whose seams are verified. Determinism reviewers must read the referenced tests before calling DE-007 on a `missing` layer; changes that add seam coverage should either place it in the canonical layer suites or extend the inventory's classifier so future reviews read the coverage from the evidence command.
