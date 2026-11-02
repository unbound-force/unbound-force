---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:18Z
identity: documentation-gate-20261101T233618-jay-flowers
tier: draft
---

When a change extends a contract that sibling changes' documentation trackers already cover (parent docs-gate issue, website docs issue, epic blog tracker), the extending change's docs-gate issue avoids tracker proliferation by explicitly partitioning scope: parent-owned per-file dispositions are referenced, not duplicated; the website-side need folds into the existing website issue with a forward-accuracy note when that issue's wording predates the extension; and residual decisions are explicitly coupled to the sibling issue's open questions so they resolve together. Verified on spawn-subtask-classification-params (unbound-force/unbound-force#834, 2026-10-26): its overlap check against #791/#742/#830/website#317 found no duplicates; its per-file dispositions match the change's recorded task-4.2 gate record; the CHANGELOG entry is present (CHANGELOG.md:828); and its website line resolves to open website#317 with a note that #317's Guard-check bullet predates the change and should document the spawn tool's parameter parity (unbound-force/replicator#143) when written — the extension folded into the existing tracker instead of a duplicate filing.
