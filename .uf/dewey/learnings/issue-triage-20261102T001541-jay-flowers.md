---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:41Z
identity: issue-triage-20261102T001541-jay-flowers
tier: draft
---

Documentation-gate issues for shipped workflow changes can be triaged definitively by verifying the gap→change→docs chain and bounding the documentation surface with targeted searches. Issue #814 (unbound-force/unbound-force) demonstrated the pattern: the docs issue named its originating change (#811, branch quick/address-feedback-diff-path-producer), the implementing PR #816 reciprocally cited it ("Documentation gate satisfied via docs issue #814", "Fixes #811"), and repo greps verified every claim — §1.4 "Fetch and Save the PR Diff" present at line 170 of both .opencode/commands/uf.address-feedback.md and its scaffold mirror, CHANGELOG entry present with the renumbering note, while docs/ (where usage.md and architecture.md describe the workflow) returned zero matches for the producer tool, proving the requested documentation work was real and pending rather than hypothetical. Quick-tier command changes still require the documentation gate: the quick tier waives spec ceremony, not documentation obligations (docs issue + CHANGELOG). Procedural workflow corrections (single tool call, section renumbering, STOP conditions) do not warrant blog or tutorial tracking — the docs issue is the proportionate content-pipeline vehicle.
