---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T19:30:19Z
identity: determinism-20261115T193019-jay-flowers
tier: draft
---

A tool-enforced gate is only as mechanical as its readable input. When a spec requires a tool to validate that a reference "corresponds to a recorded decision", the spec must name the persisted record the tool reads. In the unbound-force plugin suite, every gate validation reads files (progress state, spec-review gate records, dispatch artifacts) and tool factories receive only (exec, cwd) with no session access — so sources named as "session context" or "delegation input" are unreadable by the validating tool and force implementers to either fail closed on everything or trust caller-supplied strings, the latter re-opening the bypass the gate was built to close.
