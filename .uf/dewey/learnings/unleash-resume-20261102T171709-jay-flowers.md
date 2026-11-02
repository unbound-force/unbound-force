---
tag: unleash-resume
author: jay-flowers
category: pattern
created_at: 2026-11-02T17:17:09Z
identity: unleash-resume-20261102T171709-jay-flowers
tier: draft
---

Pattern from resuming the opsx/split-host-resolution-failure-classes pipeline (2026-11-02, fifth re-delegation after timeouts): when uf_unleash_detect_resume's artifact probe disagrees with durable checkpoint state and the orchestrator's handoff, do not follow resume_from blindly and do not immediately re-run the named step — first re-verify the probe's own freshness conditions directly against the repo. In that session detect_resume reported spec_review false while get_progress listed spec-review complete and the orchestrator had finalized the approving dispatch; running the plugin's actual code path in a small bun script (importing createUnleashDetectResumeTool and defaultGitExec from .opencode/plugins/uf-workflow/index.ts, wrapping the git exec with call logging) reproduced the false result and exposed exactly which git call returned what, converting a guess into a root cause (a defaultGitExec trim bug). The session-resume guard's precedence — get_progress and parse_tasks are authoritative and re-doing completed work is unsafe — plus a demonstrated probe bug meant the correct action was to continue from the orchestrator-indicated step. Re-dispatching a four-reviewer council to satisfy a buggy probe would have wasted a full review and, because the bug never clears on re-runs while tasks.md is unstaged, would have fixed nothing. General rule: instrument the real code path before acting on a resumability conflict, and record the defect as a retrospective learning rather than patching it mid-pipeline on a review-approved branch — phase boundaries forbid diff changes after an approving code-review verdict.
