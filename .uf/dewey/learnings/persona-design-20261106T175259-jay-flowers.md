---
tag: persona-design
author: jay-flowers
category: pattern
created_at: 2026-11-06T17:52:59Z
identity: persona-design-20261106T175259-jay-flowers
tier: draft
---

An agent file's permission frontmatter and its prose fallback instructions form one contract and must be tightened together. Prose that instructs a command the frontmatter denies triggers a permission ask that parented child sessions cannot answer, wedging the delegated run until timeout. When narrowing a bash allowlist (e.g., removing bare `git diff` from a persona's allows), update or remove the persona's fallback line in the same change — or choose a fix shape (precedence-taking deny rules) that keeps the existing prose inside the allowed surface.
