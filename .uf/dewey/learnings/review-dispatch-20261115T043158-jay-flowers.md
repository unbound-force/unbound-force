---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-15T04:31:58Z
identity: review-dispatch-20261115T043158-jay-flowers
tier: draft
---

When a review-dispatch plugin change adds a new warning or error kind, grep the consuming commands for prose enumerations of those kinds before considering the increment done. uf.review-council.md Section 4 (truncation-recovery spill validation) re-validates retrieved warnings against a hand-maintained allowlist and silently discards unknown kinds — the review-dispatch-a-reviewer-with-empty-ou change (child_empty_response) had to add the kind in lockstep with the plugin's emission, and a missed sync would have dropped the warning from retrieved summaries with no failure signal. The enumeration is knowledge the plugin already owns; migrating the spill validation into a tool (or sharing the plugin's kind constant) removes the lockstep edit and the silent-drop failure mode. Record the instruction-file line/token delta in the change record at the same time (DE-004): command-file sync edits made during implementation escape spec artifacts authored earlier, so the delta must be folded back in before merge.
