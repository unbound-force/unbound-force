---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:58:44Z
identity: uf-unleash-20261102T045844-jay-flowers
tier: draft
---

Root-caused plugin gotcha from the /uf.unleash resume run on branch opsx/review-pr-consolidate-dispatch (2026-11-02): uf_unleash_detect_resume reported the spec-review step incomplete even though spec review had completed with an approving verdict (APPROVE WITH ADVISORIES, correlation 3179b3c6-a516-4ad0-98f2-708af6a18dfe, 2 iterations). The root cause is in defaultGitExec in .opencode/lib/uf-workflow-types.ts: it applies .trim() to raw git command output before parsing. Git status --porcelain encodes each entry's status in the first two characters ('XY <path>'), so trimming the whole output eats the leading space of the first line whenever its X column is blank (for example ' M path'), and porcelainEntryPath then mis-slices the entry path. The mis-parse makes specArtifactsDirty false-positive on a clean feature directory, which poisons the spec-review freshness probe and reports the step incomplete. The bug is pre-existing and remains unfixed as of this run — fixing it is outside the current change's scope (phase boundary). Workaround for future resumes: when the probe reports spec-review incomplete but durable checkpoint state records an approving verdict, manually verify the artifact chain (artifact branch matches the current branch, base-to-head ancestry holds, and the only tasks.md delta since the review is checkbox flips), then treat the review as done; do not re-run the review on the strength of the false-positive alone, and do not fix the plugin mid-change.
