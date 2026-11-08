---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:57:24Z
identity: agent-permissions-20261108T185724-jay-flowers
tier: draft
---

A guard test pinned to one exemplar file enforces nothing for the rest of its class. PR 959 hardened divisor-linguist.md's bash allow-list with terminal metacharacter denies and shipped a guard that reads only that file — while the same PR created divisor-determinism.md without the denies, and the guard stayed green. When a guard encodes an invariant over a set of files (personas with bash allow entries), enumerate the set inside the test — glob the directory, parse each frontmatter, assert the invariant per file — instead of hardcoding the file that motivated the guard. Before extending such a guard, sweep the class for pre-existing violations so the extension lands with fixes or an explicit baseline rather than a surprise red suite (divisor-curator.md carries gh allows with no terminal denies; divisor-determinism.md also fails the wrapper-routing assertion once enumerated). Keep the enumeration predicate precise: 'bash permission is a map containing allow entries' — scalar `bash: deny`/`true`/`false` personas are out of scope, and Object.entries on a scalar throws.
