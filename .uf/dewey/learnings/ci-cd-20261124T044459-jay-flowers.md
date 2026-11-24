---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-24T04:44:59Z
identity: ci-cd-20261124T044459-jay-flowers
tier: draft
as_of: "unbound-force@90e1e32de33ae656223b0bc8ff4187eb4ada30e0"
---

fullsend-managed scaffold workflows (files carrying the "# This file is managed by fullsend. Do not edit it directly." header, e.g. .github/workflows/prioritize.yml) still contain a dependabot-managed pinned SHA in their `uses:` line. A direct Dependabot bump of that pin (branch `dependabot/github_actions/fullsend-ai/fullsend/...`) is the sanctioned update mechanism per the repo's Reusable Workflow Pinning spec and does not conflict with the do-not-edit header — the header governs hand edits, while the pinned SHA is explicitly dependabot-managed. Reviewers should not flag such a pin bump as "editing a managed file."
