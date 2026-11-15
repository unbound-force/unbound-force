---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:49:54Z
identity: determinism-20261110T194954-jay-flowers
tier: draft
---

DE-005 detection gap: the instruction-lint IL-041 mechanical-prose hint does not fire on judgment-free tool-call choreography written as numbered steps with interleaved bash blocks, and the chain can span host tools (bash + the host file-write tool), not just repo plugin tools — DE-005 review requires reading the changed instruction file in full, not relying on linter hints. Pattern: a shell-safety remediation that splits one bash call into mktemp → host file-write → gh api --input threads temp-path state through agent memory ('shell variables do not persist between separate shell calls'), and that threading is itself the DE-005 violation; the durable fix is an intent-level tool taking the composed text as a validated parameter (shipped precedents: post_feedback_replies, file_feedback_issue, uf_finale_create_pr in .opencode/plugins/). Verified on PR #1024 (uf.triage-issue.md 4.3/4.4), where the linter reported zero IL-040/IL-041/IL-042 findings on the new choreography and the violation was visible only by reading.
