---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:45:38Z
identity: agent-permissions-20261106T174538-jay-flowers
tier: draft
---

Prefix-glob bash allow rules in read-only reviewer personas are write-capable: `"git diff*": "allow"` also matches `git diff --output=<path>`, which writes an arbitrary file and bypasses `edit: deny` — and such personas consume untrusted PR prose as their primary input. Scope bash allows to the exact read-only forms the persona's own documentation names, or pair broad allows with deny rules for write-capable flags (deny takes precedence over allow). Per Constitution Principle V, reviewer personas must operate with the minimum permissions necessary; when a persona's legitimate commands are enumerable from its own documented usage, prefer exact-form allows over prefix globs.
