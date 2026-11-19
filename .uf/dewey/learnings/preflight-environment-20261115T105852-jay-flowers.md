---
tag: preflight-environment
author: jay-flowers
category: gotcha
created_at: 2026-11-15T10:58:52Z
identity: preflight-environment-20261115T105852-jay-flowers
tier: draft
as_of: "unbound-force@703a5229426874cca8cb05a2fef617910e355c11"
evidence:
  - "Makefile"
  - "AGENTS.md"
  - ".opencode/plugins/review-dispatch/index.ts"
---

Two environmental failures routinely trip the /uf.unleash quick-tier pre-flight hard gate on this machine and must be classified environmental (never change-caused, never entering the fix loop). First, make test and make coverage fail in internal/scaffold stealth-init tests under the local default Node v26.12.0 because the review-plugins init sub-tool requires Node 20 through 24 — run the Go targets with nvm's Node 24.11.1 on PATH (PATH="$HOME/.nvm/versions/node/v24.11.1/bin:$PATH" make test); CI runs Node 22, so the PR is unaffected and the failure text itself names the environmental cause. Second, make plugin-test can trip the no-lock-shipped assertion when a concurrent opencode host writes a gitignored .opencode/package-lock.json mid-run — per AGENTS.md, remove the stray lock and re-run without deleting node_modules; the vitest suite itself passes on the first run (96.83% statements against the 94% floor), and reaching the lock assertion at all proves the suite exited zero.
