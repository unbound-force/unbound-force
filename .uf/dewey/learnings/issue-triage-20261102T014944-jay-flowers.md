---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:44Z
identity: issue-triage-20261102T014944-jay-flowers
tier: draft
---

A "to be linked" cross-repo website reference inside a docs-tracking issue does not mean the website issue is missing. Documentation-gate artifacts for a merged unbound-force change arrive as a same-window cluster — the in-repo docs issue, the unbound-force/website issue, and implementation follow-ups were all filed within ~90 seconds of each other (unbound-force/unbound-force#754, unbound-force/website#312, unbound-force/unbound-force#755 on 2026-10-19) — and only the backlink in the issue body goes stale. When triaging, verify the cross-repo issue exists (gh issue view or the Dewey GitHub index) before flagging a constitution Cross-Repo Documentation gap, and check work-item type before classifying near-simultaneous keyword-overlapping issues as duplicates (#755 is a backend-swap implementation follow-up; #754 is the docs tracker for the same change).
