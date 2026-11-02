---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:47:37Z
identity: issue-triage-20261102T014737-jay-flowers
tier: draft
---

Issue-triage pattern: when a duplicate search surfaces both a predecessor docs/spec issue and an implementation follow-up that references it, they are not duplicates. Predecessor issues that file spec/design artifacts (e.g., #754 filed the fix-triage-issue-dispatch-docs spec) have a different deliverable than follow-up issues that implement a designed task from that spec (e.g., #755 implements design D4 task 5.4). Distinguish by deliverable type (docs vs code) and by the follow-up's explicit spec/task anchor; reserve the duplicate category for two issues requesting the same deliverable.
