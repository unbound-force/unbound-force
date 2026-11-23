---
tag: meta-prompt
author: jay-flowers
category: pattern
created_at: 2026-11-15T18:22:19Z
identity: meta-prompt-20261115T182219-jay-flowers
tier: draft
---

When a spec extracts command prose into a tool, the command-integration requirements must specify the tool-failure path with the same precision as the success path. Two things must be carried into the requirement explicitly: the command's per-call-site error-envelope convention (e.g. "report the error envelope verbatim and STOP"), and any success-gated cleanup phrasing ("After the artifact is written, call clear_progress") — compressing the gate to "after the artifact" and omitting the error clause leaves the rewrite's failure behavior to implementer invention, and a state-wiping step like clear_progress can then run after a failed write. Observed in the feedback-triage-artifact-tool spec review (FTA-FR-004).
