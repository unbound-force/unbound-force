---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:22:04Z
identity: issue-triage-20261102T032204-jay-flowers
tier: draft
---

When triaging issues that propose creating new tools, verify the premise against the current tree in addition to duplicate-issue searches. Issue #666 proposed wait_for_ci and check_mergeability plugin tools to replace CI-polling and mergeability prose in uf.finale.md; the duplicate-issue search returned no candidates, but by triage time the completed finale-token-reduction change (same parent epic #659) had already shipped uf_finale_ci_poll in the uf-finale plugin — different name, different plugin, same capability — and rewritten uf.finale.md to call it, so the prose the issue cited no longer existed. Capability-level supersession is invisible to issue-level duplicate search: grep the command file for the cited prose and grep the plugins for the proposed capability before marking a tool-creation issue VALID.
