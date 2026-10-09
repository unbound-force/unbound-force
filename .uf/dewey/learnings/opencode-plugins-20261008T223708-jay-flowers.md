---
tag: opencode-plugins
author: jay-flowers
category: gotcha
created_at: 2026-10-08T22:37:08Z
identity: opencode-plugins-20261008T223708-jay-flowers
tier: draft
---

When implementing TypeScript plugin tools in the OpenCode plugin framework, the tool() function from @opencode-ai/plugin returns a complex framework-defined type that makes explicit return type annotation impractical. The review council's architect persona flagged this as a LOW convention deviation (CS-005 return type inference) but marked it as a reasonable exception. This is a recurring pattern — factory functions that return framework-wrapped types should use inference rather than trying to explicitly annotate the complex generic return type, avoiding fragile coupling to framework internals.
