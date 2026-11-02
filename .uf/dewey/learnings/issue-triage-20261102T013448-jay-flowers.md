---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:34:48Z
identity: issue-triage-20261102T013448-jay-flowers
tier: draft
---

Duplicate assessment for content-pipeline issues must compare content type and label, not subject keywords alone. Triaging tutorial issue #767 ("tutorial: recovering a stranded review dispatch") in unbound-force/unbound-force, the primary keyword search ("tutorial stranded review dispatch recovery correlation_id workflow") surfaced only #766 ("docs: review-dispatch recovery tools and consolidation changes"). Both track the same underlying change (fix-review-dispatch-recovery), but #766 is a docs-labeled reference-documentation issue while #767 is a tutorial-labeled operator walkthrough with a distinct audience, structure, and prerequisites. Keyword overlap alone would misclassify them as duplicates. The discriminating fields are the label (docs vs blog vs tutorial) and the deliverable format; when subject overlap exists across content types, cross-reference the issues during execution rather than closing one as a duplicate.
