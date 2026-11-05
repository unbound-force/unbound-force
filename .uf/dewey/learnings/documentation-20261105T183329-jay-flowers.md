---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:33:29Z
identity: documentation-20261105T183329-jay-flowers
tier: draft
---

When a reference table documents bounds for sibling tool parameters, verify each column against the schema that validates that specific parameter. In PR #950 (delegation-timeout-budgets), docs/configuration.md copied idle_timeout's enforced 60,000 ms floor into the timeout column, whose zod schema is positive-integer-with-ceiling only; the spec (DT-FR-003) and design doc claimed ceilings only, so the floor was a documentation invention. Parameter defaults and sibling-parameter floors are frequent sources of this drift. (From divisor-curator review of PR #950.)
