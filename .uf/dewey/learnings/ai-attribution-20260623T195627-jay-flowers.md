---
tag: ai-attribution
author: jay-flowers
category: gotcha
created_at: 2026-06-23T19:56:27Z
identity: ai-attribution-20260623T195627-jay-flowers
tier: draft
---

When adding a model name to AI attribution trailers in git commits, the spec review council identified two critical security requirements: (1) the model name must be validated against `[a-zA-Z0-9._-]+` before insertion into the git trailer to prevent trailer injection (newlines, colons, or shell metacharacters in the model name could inject arbitrary trailers), and (2) the extraction algorithm must be explicitly specified (strip after last `/`, strip after first `@`) rather than left vague as "derive from system configuration." The Adversary reviewer specifically confirmed that the `[a-zA-Z0-9._-]` allowlist prevents git trailer injection, shell metacharacter injection, and newline injection. The `unknown-model` literal also passes this validation.
