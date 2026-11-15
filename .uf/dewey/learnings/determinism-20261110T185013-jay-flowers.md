---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-10T18:50:13Z
identity: determinism-20261110T185013-jay-flowers
tier: draft
---

When a security fix forces splitting a single shell block into a multi-call choreography (shell mktemp → host file-write tool → shell gh api with a transcribed `<PLACEHOLDER>` path), that split is the signal the sequence is a DE-005 intent-level tool candidate: every step after the first is fully determined by the previous step's output, and the only judgment product (the composed payload) becomes a tool parameter. Shipped precedents: file_feedback_issue and post_feedback_replies encapsulate temp-file creation, permissions, gh api --input posting, and cleanup inside the tool. File the extraction follow-up in the same change record so the interim prose form does not silently become the mechanism of record.
