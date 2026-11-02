---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:19Z
identity: issue-triage-20261102T033519-jay-flowers
tier: draft
---

Triage corollary for tool-extraction issues that structuralize human gates or security-relevant routing constants (from unbound-force #664, post_github_review): converting a prose-instructed MANDATORY GATE into a required tool parameter with fail-closed refusal (gate_confirmed) both preserves and strengthens the gate — it becomes structural, survives context compression, and must be re-supplied on session resume — but only if the acceptance criteria pin the refusal path with an explicit negative test (gate_confirmed false or absent → tool refuses to post). Happy-path contract tests alone pass green while silently allowing a regression back to ungated posting; the #669 gate-preservation gotcha generalizes from prose gates to structural gates. The same fail-closed discipline applies to internalized mappings: a verdict→GitHub-event mapping absorbed from a sibling issue (#671) should reject unknown verdict strings at validation rather than defaulting, and the single gh api --input posting call should be pinned by tests as the atomicity guarantee for review-plus-inline-comments.
