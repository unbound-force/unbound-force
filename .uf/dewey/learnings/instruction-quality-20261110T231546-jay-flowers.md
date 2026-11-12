---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:15:46Z
identity: instruction-quality-20261110T231546-jay-flowers
tier: draft
---

When a tool's output contract changes, grep every command, skill, and agent file for the retired contract wording before merging: contract sentences get restated in several sections (preview description, lesson storage, failure recovery), and every surviving restatement becomes a contradiction a future reader executes. Verifying the full-lesson-proposals change needed two greps — the retired bound (512) and the contract phrase (capped preview) — across .opencode/commands/ and .opencode/skills/ to prove the two edited sentences were the only carriers of the old claim; the neighboring consumers (uf.address-feedback.md, uf.triage-issue.md, uf.unleash.md) described the preview without the cap. Conclusion: retire a contract claim with a targeted grep of its exact literals, never from memory of where it was written.
