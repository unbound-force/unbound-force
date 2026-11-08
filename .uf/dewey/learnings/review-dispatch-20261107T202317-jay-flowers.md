---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:17Z
identity: review-dispatch-20261107T202317-jay-flowers
tier: draft
---

Registering a repo-local reviewer persona in .uf/reviewer-capabilities.yaml requires mapping the persona's declared file domain through the mechanical path classifier (.opencode/lib/review-diff-manifest.ts classifyPathHints), because dispatch eligibility is strict scope-intersection: review-dispatch excludes a reviewer with code "scope-miss" when none of its scopes intersect the change categories. The taxonomy does not follow intent: `.opencode/commands/*.md` classifies cli-ux (the cli-ux pattern deliberately takes precedence over the documentation pattern for command .md files), while `.opencode/agents|skills|skill|uf/packs/*.md` and AGENTS.md classify documentation, and `.opencode/lib/*.ts` classifies standard. A persona whose declared domain includes command prose (PR #958's divisor-linguist, registered scopes: [documentation]) is therefore never dispatched for command-only changes — the judgment layer silently skips the surface the persona was built to guard, even though the deterministic pre-flight layer (whose paths list is authored independently) still covers it. When adding a reviewer entry, enumerate every path pattern the persona claims, classify each, and include every resulting category in scopes; scoping only [standard] to a domain-specific persona over-fires it on every code change, while omitting a category under-fires it.
