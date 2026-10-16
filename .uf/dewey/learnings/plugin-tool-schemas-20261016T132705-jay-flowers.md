---
tag: plugin-tool-schemas
author: jay-flowers
category: pattern
created_at: 2026-10-16T13:27:05Z
identity: plugin-tool-schemas-20261016T132705-jay-flowers
tier: draft
---

Schema-exposure regression tests for OpenCode plugin tools that assert only Object.keys(tool.args) key-name equality guard parameter presence but not the full declaration contract: an optional-to-required optionality regression passes (a required zod-inferred field is assignable to an optional-annotated execute parameter, so tsc stays silent; only required-to-optional fails typing), and description drift via a dropped .describe() is unguarded. For full declaration-contract coverage, also assert zod introspection per field — e.g. expect(args.setUpstream.isOptional()).toBe(true), expect(args.prNumber.isOptional()).toBe(false) — mirroring whatever parameter-fidelity requirements the declaration migration must preserve.
