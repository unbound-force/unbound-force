---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:15:49Z
identity: review-dispatch-20261102T001549-jay-flowers
tier: draft
---

Triage/docs precision for the review-dispatch correlation lifecycle (unbound-force #808): the correlation_id parameter contract is asymmetric across the tool family. dispatch_status and consolidate_dispatch treat correlation_id as OPTIONAL with latest-wins default resolution (invoked without an explicit ID, they resolve the latest dispatch session; with an explicit ID, an earlier session). The recovery tools redispatch_failed_runs and discard_orphaned_submissions REQUIRE an explicit correlation_id as their first parameter — recovery operators run from a fresh parent session after a restart, when no in-process dispatch state exists. Checklists or docs that lump all four tools under "accept an explicit correlation_id" (as #808's recovery-tools checklist item does) are directionally correct but understate the recovery contract; documentation produced from that checklist must preserve the asymmetry so readers do not expect a latest-wins default where the parameter is mandatory. Verified against the tool schemas in the parent-session context and the CHANGELOG recovery-tool signatures (discard_orphaned_submissions(correlation_id, agents?), redispatch_failed_runs(correlation_id, agents?, timeout?)) on 2026-11-01.
