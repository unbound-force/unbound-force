---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-20T18:19:24Z
identity: spec-authoring-20261120T181924-jay-flowers
tier: draft
as_of: "unbound-force@581834811fb8f5636c9a65c632b8d2d812b24fac"
evidence:
  - "dewey:.uf/artifacts/dispatch/bedd15b4-b800-40c3-8671-8eb747ed9481.json"
---

When an OpenSpec requirement delegates semantics to an existing protocol by name ("existing review-context Protocol 2 parsing"), the delegation must pin cardinality and selection, not just identity and priority. Verified while reviewing intent-satisfaction-check (2026-11-20): RC-FR-010 resolved "the originating issue reference" (singular) from PR-body parsing whose actual contract is set-valued — review-context Protocol 2 validates and fetches up to 5 linked issues and extracts criteria "from each fetched issue body" — so multi-reference PR bodies were left undefined while every downstream consumer (RC-FR-011/012, task 3.1 field shapes) assumed exactly one issue. Pattern: before writing "the <singular noun> from <protocol>", read the delegated protocol's output shape and either state the within-source selection rule (first reference wins / all references grounded and evaluated) or make the plural explicit, and add a scenario for the multi-candidate input.
