---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-04T20:08:44Z
identity: review-dispatch-20261104T200844-jay-flowers
tier: draft
---

When a review council runs against a working tree checked out at the PR base SHA, grep/read results for diff-touched files show the BASE state, not the head. Classify every grep hit by whether its file appears in the PR's changed-file list before treating it as a stale reference: hits inside diff-touched files are pre-change text the PR replaces (not findings), while hits in unchanged files are the only real stale-reference evidence. In PR #957 review (stop-shipping-plugin-lock), repo-wide greps for "npm ci  [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
