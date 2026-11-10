---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:21:23Z
identity: instruction-lint-20261110T152123-jay-flowers
tier: draft
---

During PR review dispatches, a clean `instruction-lint --json --base <sha> --changed` run is change-scoped evidence ONLY when the working tree is at the PR head. When the checkout sits on main (or any ref that does not carry the head changes), the linter computes its delta from local HEAD: delta.files comes back empty (or reflects unrelated drift), findings are zero, and the pre-flight scope check skips the same way ("no in-scope files") even though .uf/preflight.yaml lists AGENTS.md in the check's path scope. Verify the checkout with git status first; if the head SHA is missing locally (git diff <base> <head> returns "fatal: bad object"), disclose in the review summary that Step 1 produced no evidence about the PR's changed instruction lines and let the judgment review carry the finding load. Do not hand-run the linter's checks against the diff text. Observed reviewing PR #1005 (AGENTS.md instruction change) from a main checkout 144 commits ahead of origin/main.
