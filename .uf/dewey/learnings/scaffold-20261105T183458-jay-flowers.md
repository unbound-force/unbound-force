---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:34:58Z
identity: scaffold-20261105T183458-jay-flowers
tier: draft
---

Adding a path to `.gitignore` does not untrack a file that is already tracked in git — the tracked copy keeps showing dirty on every write while appearing ignored. When a scaffolding tool moves a file from tool-owned to ignored (stop-shipping-plugin-lock moving `.opencode/package-lock.json`), existing deployments keep a stale tracked copy that the ignore line cannot clean up, and external writers (opencode) keep churning it. The complete pattern is three parts: stop writing/reading the file, ignore it for fresh checkouts, and document an explicit migration (git rm --cached) for existing deployments.
