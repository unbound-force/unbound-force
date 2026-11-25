---
tag: testing
author: jay-flowers
category: gotcha
created_at: 2026-11-25T00:22:10Z
identity: testing-20261125T002210-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

When stubbing a CLI's JSON output in tests, seed the payload shape the CLI actually emits, not a simplified equivalent. In the intent-satisfaction-check changeset, every gh issue view stub returned labels as plain strings, while real gh returns [{name, color, description}] objects — the production parser deliberately supported both forms (documented in its doc comment), so all tests passed while the only branch real-world traffic exercises (object-form labels, which also drives the intake-kit strict-grounding tier detection) had zero coverage. Grep for the fixture shape against the real producer's documented output before trusting a green suite.
