---
tag: uf-workflow-resume
author: jay-flowers
category: gotcha
created_at: 2026-11-02T06:09:00Z
identity: uf-workflow-resume-20261102T060900-jay-flowers
tier: draft
---

During the opsx/host-fallback-provenance /uf.unleash resume (2026-11-02, the issue #901 pipeline), uf_unleash_detect_resume reported spec_review false while every input of its documented specs-mode freshness rule verified true: the most recent current-branch specs-mode artifact under .uf/artifacts/dispatch (correlation eac48ce9, verdict APPROVE WITH ADVISORIES) recorded head SHA fa51ebed, an ancestor of current HEAD 9e1ff855 (git merge-base --is-ancestor exits 0), and git status --porcelain over the feature directory showed only tasks.md checkbox edits, which the rule explicitly excludes. Re-executing the documented algorithm directly — a node simulation of parseReviewDispatchArtifact and mostRecentArtifact over the real artifact store, plus the two git probes — proved the on-disk uf-workflow plugin code computes specReviewDone true, so the running tool instance's false was stale-tool behavior inconsistent with its own contract (code mode passed only because its recorded head SHA equaled HEAD exactly). The working rule: when a resume-detection tool's output contradicts its documented freshness contract, do not blindly re-run the blocked gate — re-verify each documented input against the durable artifacts (parse the store, run the git probes yourself) and let artifact evidence plus checkpoint claims plus explicit human instruction decide. This avoided a full five-agent council re-review of already-approved, committed, unchanged spec artifacts.
