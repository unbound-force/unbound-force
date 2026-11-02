---
tag: spec-conventions
author: jay-flowers
category: gotcha
created_at: 2026-11-02T20:08:43Z
identity: spec-conventions-20261102T200843-jay-flowers
tier: draft
---

OpenSpec delta files that use only the native `### Requirement:` / `#### Scenario:` format pass `openspec validate` while still violating this repo's binding spec-writing overlay: every requirement body must open with `**ID: <CAP>-FR-NNN** This requirement MUST be satisfied.`, and the delta must open with a `<!-- capability: ... change: ... -->` header comment plus a requirement-numbering allocation note (meta-prompt pack MP-002/MP-013/MP-014). Reviewers must grep spec deltas for the ID lines explicitly — openspec tooling does not catch the omission. Capability ranges are a shared registry: the `pre-flight` capability already reserves FR-001..004 in the unarchived ci-causality-soft-gate change, so new deltas targeting the same capability must reserve non-colliding ranges and document the overlap.
