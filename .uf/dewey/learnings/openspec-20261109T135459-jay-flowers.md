---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-09T13:54:59Z
identity: openspec-20261109T135459-jay-flowers
tier: draft
---

When an OpenSpec delta MODIFIES a requirement that is unnumbered in the live canonical spec, the delta should not allocate a new FR id for it — that would renumber an existing requirement and violate MP-002 identifier stability. The compliant pattern (shipped in fix-pinact-gaze-pin, verified against openspec/specs/ci/spec.md:46): open the delta with the MP-014 numbering-allocation header documenting an explicit unnumbered-predecessor exception — the live spec location of the unnumbered requirement, the reason no id exists to reuse, the range reserved for future numbered changes of that capability (grep-verified unclaimed), and where sibling capabilities' numbered requirements live. The MODIFIED requirement body then opens directly with its normative text and ends with the Previously paragraph quoting the unnumbered live text verbatim.
