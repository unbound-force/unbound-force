---
tag: supply-chain
author: jay-flowers
category: decision
created_at: 2026-11-10T15:21:08Z
identity: supply-chain-20261110T152108-jay-flowers
tier: draft
---

When a manifest-merge feature widens an npm install surface to user-declared packages, `--ignore-scripts` on the activation install is the load-bearing mitigation: it blocks install-time code execution from repo-controlled manifest entries, leaving only import-time execution — a capability repo-controlled plugin code already had, so the marginal trust-boundary change is near zero. Shipped security pins must be forced (overlaid), never merged, so the user-controlled file cannot downgrade them, and security overrides must be restored tool-owned outside the merged block.
