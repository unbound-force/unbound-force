---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:26Z
identity: issue-triage-20261102T023726-jay-flowers
tier: draft
---

Documentation-gap issues for shipped behavior changes are triage-verifiable with two mechanical checks before accepting the claim on faith: (1) grep the command/source files for the claimed behavior to confirm it is implemented as described — issue #717 claimed the three review commands each finalize via exactly one provenance-bearing consolidate_dispatch call and prohibit manual payload assembly and direct finalize_review_dispatch use; all three commands carried that instruction verbatim, and the claimed #712 gh-cli fix was confirmed by the absence of any --slurp reference in .opencode/commands/; (2) grep the user-facing docs tree for the feature name to confirm the coverage gap is real — docs/ contained zero user-facing consolidate_dispatch references. Implemented-behavior confirmation plus demonstrated docs absence converts a subjective "docs needed" request into objective VALID evidence; the spec's tasks.md (task 4.4, "file a documentation issue before PR merge") additionally proved the issue itself was the documentation-gate artifact rather than an ad-hoc request.
