---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:13:32Z
identity: spec-governance-20261107T171332-jay-flowers
tier: draft
---

When a spec mandate of the form "MUST keep the X that the file already declares" is implemented by merging against a byte-exact structural anchor (e.g., the literal `\n  "overrides": {\n` in a tool-owned template), the anchor-miss path silently violates the mandate: the merge no-ops and the user's declared entries drop out of their git-tracked file with no error, warning, or downstream check. In PR 957's preservePluginDevDependencies, the assessment framing claimed the passthrough was "documented D2a design intent," but retrieving the artifacts showed design.md D2a documents only the re-insertion — the word "anchor" appears nowhere in the change, the spec delta carves no exception, and the branch has no justifying comment; the only artifact pinning the passthrough was the unit test. A test asserts behavior but is not a ratified design rationale. Governance pattern for such fallbacks: (a) record the boundary decision in the design doc / spec delta, (b) carry a justifying comment at the branch (default.md CS-006), (c) emit a runtime signal (warning naming the file and reason) whenever the mandate is silently unmet — and remember no downstream check catches the drop when the dropped entries are also undeclared for every later validator (doctor's declared-to-installed comparison skips them). Failing the whole operation is a separate availability trade-off the author may legitimately decline; the signal is the non-negotiable part.
