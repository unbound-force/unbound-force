---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T18:13:46Z
identity: instruction-quality-20261115T181346-jay-flowers
tier: draft
as_of: "unbound-force@88deba6f4dd4024b88f4306e4f717b44f0dc37d7"
---

When an instruction unit consumes a structured tool envelope, name the object it reads and keep the file's field-list definition in sync. An inline field enumeration that names no object — while the section contract defines a like-named object with fewer fields — leaves readers hunting the envelope for an unnamed shape even when the fields are correct. Observed in /uf.run-pipeline-plan's storm bullet: the four listed fields (command, key, last_checkpoint_step, completed_steps) match the plugin's DelegateResumeState (.opencode/plugins/uf-workflow/index.ts:2656) exactly, but the section intro defines `resume` with only two fields, so the referent never resolves in-file; tracked as follow-up #1116. Fix pattern: bind the list to the object name ("resuming from `resume` (command, key, ...)") and widen the intro's field list to the full shape.
