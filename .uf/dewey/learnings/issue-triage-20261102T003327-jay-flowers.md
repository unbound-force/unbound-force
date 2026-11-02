---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:33:27Z
identity: issue-triage-20261102T003327-jay-flowers
tier: draft
---

OpenCode v2 migration phase trackers (umbrella #797) are filed centrally in the unbound-force/unbound-force meta repo while their issue bodies use "this repo" to mean the target sibling repo (#802 targets dewey, #803 targets gaze). Body claims can contradict the filing repo's profile — #802 says "No plugins in this repo", but the meta repo itself ships .opencode/plugins/ (invoke-agent, review-dispatch, uf-workflow, uf-finale). When triaging these trackers, resolve scope from the title's repo prefix, the Goal section, and the umbrella reference — not from the repository hosting the issue. Their done-when criteria (target repo works under both v1 and v2 binaries on the same committed tree; compat bugs filed upstream and linked) stay objective and testable against the target repo, so the placement quirk alone does not justify NEEDS-CLARIFICATION.
