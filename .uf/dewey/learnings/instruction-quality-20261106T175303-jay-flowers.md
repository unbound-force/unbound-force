---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-06T17:53:03Z
identity: instruction-quality-20261106T175303-jay-flowers
tier: draft
---

An agent file's bash permission allowlist and its body prose form one contract. When a permission rule is narrowed, every command the body instructs the agent to run — such as a no-caller-diff fallback like "identify changes with `git diff` and `git status`" — must either remain allowed or be updated in the same change; otherwise the instruction routes the agent into the permission gate with no stated failure handling. Review permission-frontmatter changes side by side with the body's documented command invocations, and extend the file's graceful-degradation table when a documented command can now be denied.
