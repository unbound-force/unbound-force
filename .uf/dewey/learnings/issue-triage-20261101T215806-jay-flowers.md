---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:58:06Z
identity: issue-triage-20261101T215806-jay-flowers
tier: draft
---

A documentation-tracking issue is triage-valid in the unbound-force meta repo when it carries three completeness signals: (1) it enumerates the operator-visible behavior changes rather than only referencing the implementation issue, (2) it states what the documentation must cover so a docs author can act without re-deriving scope, and (3) it explicitly disambiguates scope from sibling issues covering the same component. Issue #880 (docs: manifest + on-demand diff retrieval for review dispatch) demonstrates the pattern: it enumerates the five user-visible facets of the change, links the originating issue #877, the spec directory, and the sibling recovery-docs issue #766, and every capability claim is verifiable against the implemented tool surface and the completed spec tasks — which made duplicate classification and validity assessment objective rather than judgment-based.
