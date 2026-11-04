---
tag: tool-input-validation
author: jay-flowers
category: gotcha
created_at: 2026-11-03T16:46:06Z
identity: tool-input-validation-20261103T164606-jay-flowers
tier: draft
---

OpenCode plugin hosts do not enforce a tool's declared Zod arg shape before execute runs: structurally invalid arguments (e.g., checkpoint `phase: 5` where an object {number, verdict, inventory} is declared) reach execute and surface as misleading derived values ("requires step implement-phase-undefined", issue #936). Fix pattern: extract the arg schema as one shared constant, re-validate inside execute through the plugin's parseToolInput helper, and return INVALID_INPUT naming the expected shape. The decla
