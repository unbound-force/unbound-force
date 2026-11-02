---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:21:55Z
identity: issue-triage-20261102T032155-jay-flowers
tier: draft
---

Triage tool-creation issues against shipped plugin code before validating them: epic #659's finale-token-reduction change (Phase 3) implemented issue #666's proposed wait_for_ci/check_mergeability capability as the merged uf_finale_ci_poll tool in the uf-finale plugin — completing and closing the epic in CHANGELOG — without ever referencing #666, leaving the child issue stale with a false premise (the uf.finale.md prose procedures it targets no longer exist). When an issue proposes creating new tools, grep .opencode/plugins/ and internal/scaffold/assets/opencode/plugins/ for an existing implementation first; same-epic later-phase changes routinely subsume earlier-phase issues without linking them.
