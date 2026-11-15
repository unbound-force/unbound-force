---
tag: untrusted-input
author: jay-flowers
category: pattern
created_at: 2026-11-15T05:06:25Z
identity: untrusted-input-20261115T050625-jay-flowers
tier: draft
---

Untrusted-input markers are scoped per consumption point, not per data class. In uf.review-pr.md, Step 3.12 marks the prior review body as data-never-instructions for the light path, but PR 1077's new Step 8a branch — which parses the same prior review body for a staleness comparison in both review modes — carried no marker. When adding any step that reads fetched external text (review bodies, issue text, diffs), carry the established marker to the new location even when sibling steps already mark the same data class elsewhere; the file's own convention treats a missing marker at a new consumption point as a HIGH instruction-quality finding.
