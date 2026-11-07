---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:09:37Z
identity: supply-chain-20261105T160937-jay-flowers
tier: draft
---

When a change stops tracking a lockfile, enumerate every detective control that consumed it — PR-time scanners, scheduled newly-disclosed-CVE scans, dependency-graph alert feeds, and install-time npm audit — and test whether the residual named in the risk acceptance actually covers the loss: "package.json remains scannable" covers only exact-pinned direct dependencies, never the transitive tree. A throwaway-lock scan restores transitive visibility without reintroducing a tracked lock and composes with the change's own guards: generate with the project's working npm pin via `npm install --package-lock-only --ignore-scripts` (--package-lock-only exercises the same Arborist ideal-tree build that crashes older bundled npm on complex peer graphs, so reuse the pinned npm), scan with osv-scanner, discard — the gitignore rule that keeps the lock untracked also makes the ephemeral CI lock commit-safe. Note that `--no-audit` removes install-time advisory signal that does not depend on a lock at all, so it compounds, rather than merely accompanies, lockless scanning loss.
