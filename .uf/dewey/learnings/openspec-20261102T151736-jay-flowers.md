---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T15:17:36Z
identity: openspec-20261102T151736-jay-flowers
tier: draft
---

OpenSpec MODIFIED entries must reproduce the predecessor requirement's id line; when the predecessor requirement is unnumbered (created before the allocation-header convention), the successor delta's allocation header must list it as unnumbered alongside the tool-requirement exception — otherwise the delta's numbering map claims completeness while its own MODIFIED content contradicts it. Observed in split-host-resolution-failure-classes: the Advisor Host-Model Fallback requirement has stayed unnumbered across three delta generations because each successor copied the body shape without normalizing identity or documenting the exception.
