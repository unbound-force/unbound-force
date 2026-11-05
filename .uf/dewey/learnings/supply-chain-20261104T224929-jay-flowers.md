---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:49:29Z
identity: supply-chain-20261104T224929-jay-flowers
tier: draft
---

Removing a lockfile to fix multi-writer conflicts (uf and opencode both installing into .opencode/) trades away three controls at once: transitive version pinning, lock-scanner input (OSV-Scanner/Dependabot), and — when --no-audit is added for quiet installs — npm's install-time advisory report. Review pattern for such changes: audit every install path, not just the primary activation path, for the design's stated mitigations (in PR #957 --ignore-scripts traveled with the uf init activation path but not the
