---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T11:37:35Z
identity: instruction-quality-20261115T113735-jay-flowers
tier: draft
---

Instruction prose that branches on tool-rendered failure envelopes must cite fields through the envelope's declared object name, and a section's field enumeration must be extended when a new bullet cites fields beyond it. In .opencode/commands/uf.run-pipeline-plan.md's Delegation timeout recovery section, the contract sentence still declares `resume` as (`last_checkpoint_step`, `completed_steps`) while the tool-rendered resume object (readDelegateResumeState in .opencode/plugins/uf-workflow/index.ts) carries four fields (`command`, `key`, `last_checkpoint_step`, `completed_steps`); the storm bullet added in this change cites all four through a coined phrase ("the envelope's durable progress state"), leaving the section with three names for one object. When reviewing or writing such bullets, verify referents against the tool's envelope-construction code, not only against sibling prose.
